package file

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/rate"
)

func NewJsonDb(runPath string) *JsonDb {
	return &JsonDb{
		RunPath:          runPath,
		TaskFilePath:     filepath.Join(runPath, "conf", "tasks.json"),
		HostFilePath:     filepath.Join(runPath, "conf", "hosts.json"),
		ClientFilePath:   filepath.Join(runPath, "conf", "clients.json"),
		GlobalFilePath:   filepath.Join(runPath, "conf", "global.json"),
		ApiTokenFilePath: filepath.Join(runPath, "conf", "api_tokens.json"),
		SqliteFilePath:   filepath.Join(runPath, "conf", "nps.db"),
	}
}

type JsonDb struct {
	Tasks            sync.Map
	Hosts            sync.Map
	HostsTmp         sync.Map
	Clients          sync.Map
	ApiTokens        sync.Map
	Global           *Glob
	RunPath          string
	ClientIncreaseId int32  //client increased id
	TaskIncreaseId   int32  //task increased id
	HostIncreaseId   int32  //host increased id
	ApiTokenIncreaseId int32 //api token increased id
	TaskFilePath     string //task file path
	HostFilePath     string //host file path
	ClientFilePath   string //client file path
	GlobalFilePath   string //global file path
	ApiTokenFilePath string //api token file path
	SqliteFilePath   string //sqlite db file path (phase 0+ of JSON->SQLite migration)
	// dirty tracking (Phase 8 / P1 optimisation): per-type sets of IDs
	// that have been mutated outside CRUD (flow counters, online state).
	// Periodic flushers (server.flowSession) only UPSERT entries listed
	// here, eliminating full-table scans and unnecessary fsyncs in the
	// steady state. CRUD paths still write through synchronously.
	dirtyHosts   sync.Map // key: int (host.Id) -> struct{}{}
	dirtyTasks   sync.Map // key: int (tunnel.Id) -> struct{}{}
	dirtyClients sync.Map // key: int (client.Id) -> struct{}{}
	dirtyGlobal  atomic.Bool
	// SQLite is the opaque handle for the optional SQLite store. Type is
	// `any` so that this package does not import the sqlite driver,
	// keeping `npc` (which embeds `lib/file` only for its data models)
	// free of the modernc.org/sqlite + libc transitive dependency.
	// The server entrypoint (cmd/nps) populates this with a
	// *lib/file/sqlitedb.Store at startup.
	SQLite any
	// ClientStore, when non-nil, makes Client CRUD dual-write to a
	// persistence backend (typically SQLite). The in-memory sync.Map
	// remains the runtime registry; this hook only persists the
	// admin-managed subset of fields. Populated by cmd/nps after the
	// SQLite store is opened. nil for npc and during early startup.
	ClientStore ClientPersister
	// TaskStore is the analogous hook for the Tasks (Tunnel) table.
	TaskStore TaskPersister
	// HostStore is the analogous hook for the Hosts table.
	HostStore HostPersister
	// GlobalStore is the analogous hook for the singleton Glob record.
	GlobalStore GlobalPersister
}

// ClientPersister is the persistence hook used by Client CRUD in db.go.
// Defined in lib/file (rather than in lib/file/sqlitedb) so that the
// dual-write logic does not pull the sqlite driver into npc.
type ClientPersister interface {
	UpsertClient(*Client) error
	DeleteClient(id int) error
}

// TaskPersister mirrors ClientPersister for the Tunnel (tasks) table.
type TaskPersister interface {
	UpsertTask(*Tunnel) error
	DeleteTask(id int) error
}

// HostPersister mirrors ClientPersister for the Host table.
type HostPersister interface {
	UpsertHost(*Host) error
	DeleteHost(id int) error
}

// GlobalPersister mirrors ClientPersister for the singleton Glob record.
type GlobalPersister interface {
	UpsertGlobal(*Glob) error
}

func (s *JsonDb) LoadTaskFromJsonFile() {
	loadSyncMapFromFile(s.TaskFilePath, func(v string) {
		var err error
		post := new(Tunnel)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		if post.Client, err = s.GetClient(post.Client.Id); err != nil {
			return
		}
		s.Tasks.Store(post.Id, post)
		if post.Id > int(s.TaskIncreaseId) {
			s.TaskIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) LoadClientFromJsonFile() {
	loadSyncMapFromFile(s.ClientFilePath, func(v string) {
		post := new(Client)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		if post.RateLimit > 0 {
			post.Rate = rate.NewRate(int64(post.RateLimit * 1024))
		} else {
			post.Rate = rate.NewRate((2 << 23) * 1024)
		}
		post.Rate.Start()
		post.NowConn = 0
		s.Clients.Store(post.Id, post)
		if post.Id > int(s.ClientIncreaseId) {
			s.ClientIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) LoadHostFromJsonFile() {
	loadSyncMapFromFile(s.HostFilePath, func(v string) {
		var err error
		post := new(Host)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		if post.Client, err = s.GetClient(post.Client.Id); err != nil {
			return
		}
		s.Hosts.Store(post.Id, post)
		if post.Id > int(s.HostIncreaseId) {
			s.HostIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) LoadGlobalFromJsonFile() {
	loadSyncMapFromFileWithSingleJson(s.GlobalFilePath, func(v string) {
		post := new(Glob)
		if json.Unmarshal([]byte(v), &post) != nil {
			return
		}
		s.Global = post
	})
}

func (s *JsonDb) LoadApiTokensFromJsonFile() {
	if !common.FileExists(s.ApiTokenFilePath) {
		return
	}
	loadSyncMapFromFile(s.ApiTokenFilePath, func(v string) {
		post := new(ApiToken)
		if json.Unmarshal([]byte(v), post) != nil {
			return
		}
		s.ApiTokens.Store(post.Id, post)
		if post.Id > int(s.ApiTokenIncreaseId) {
			s.ApiTokenIncreaseId = int32(post.Id)
		}
	})
}

func (s *JsonDb) GetClient(id int) (c *Client, err error) {
	if v, ok := s.Clients.Load(id); ok {
		c = v.(*Client)
		return
	}
	err = errors.New("未找到客户端")
	return
}

var hostLock sync.Mutex

// MarkHostDirty records that the in-memory Host with the given id has
// been mutated outside the CRUD path (typically by FlowAddHost). The
// next FlushHostsToStore tick will UPSERT only the marked rows.
func (s *JsonDb) MarkHostDirty(id int)   { s.dirtyHosts.Store(id, struct{}{}) }
func (s *JsonDb) MarkTaskDirty(id int)   { s.dirtyTasks.Store(id, struct{}{}) }
func (s *JsonDb) MarkClientDirty(id int) { s.dirtyClients.Store(id, struct{}{}) }
func (s *JsonDb) MarkGlobalDirty()       { s.dirtyGlobal.Store(true) }

// FlushHostsToStore performs an UPSERT for every Host in the in-memory
// registry against the configured HostStore (SQLite). It is a no-op when
// no store is wired (e.g. early startup or npc).
//
// Phase 7: JSON write paths have been removed; SQLite is the sole
// persistence layer. Call sites that previously wrote conf/hosts.json
// now invoke this helper instead.
//
// Phase 8 (P1): only walks the dirty set; the entries are cleared after
// a successful UPSERT so a quiescent system performs zero writes.
func (s *JsonDb) FlushHostsToStore() {
	hostLock.Lock()
	defer hostLock.Unlock()
	if s.HostStore == nil {
		return
	}
	s.dirtyHosts.Range(func(k, _ interface{}) bool {
		id, ok := k.(int)
		if !ok {
			s.dirtyHosts.Delete(k)
			return true
		}
		v, exists := s.Hosts.Load(id)
		if !exists {
			s.dirtyHosts.Delete(id)
			return true
		}
		h, ok := v.(*Host)
		if !ok || h == nil || h.NoStore {
			s.dirtyHosts.Delete(id)
			return true
		}
		if err := s.HostStore.UpsertHost(h); err == nil {
			s.dirtyHosts.Delete(id)
		}
		return true
	})
}

var taskLock sync.Mutex

// FlushTasksToStore is the Tunnel equivalent of FlushHostsToStore.
func (s *JsonDb) FlushTasksToStore() {
	taskLock.Lock()
	defer taskLock.Unlock()
	if s.TaskStore == nil {
		return
	}
	s.dirtyTasks.Range(func(k, _ interface{}) bool {
		id, ok := k.(int)
		if !ok {
			s.dirtyTasks.Delete(k)
			return true
		}
		v, exists := s.Tasks.Load(id)
		if !exists {
			s.dirtyTasks.Delete(id)
			return true
		}
		t, ok := v.(*Tunnel)
		if !ok || t == nil || t.NoStore {
			s.dirtyTasks.Delete(id)
			return true
		}
		if err := s.TaskStore.UpsertTask(t); err == nil {
			s.dirtyTasks.Delete(id)
		}
		return true
	})
}

var clientLock sync.Mutex

// FlushClientsToStore is the Client equivalent of FlushHostsToStore.
func (s *JsonDb) FlushClientsToStore() {
	clientLock.Lock()
	defer clientLock.Unlock()
	if s.ClientStore == nil {
		return
	}
	s.dirtyClients.Range(func(k, _ interface{}) bool {
		id, ok := k.(int)
		if !ok {
			s.dirtyClients.Delete(k)
			return true
		}
		v, exists := s.Clients.Load(id)
		if !exists {
			s.dirtyClients.Delete(id)
			return true
		}
		c, ok := v.(*Client)
		if !ok || c == nil || c.NoStore {
			s.dirtyClients.Delete(id)
			return true
		}
		if err := s.ClientStore.UpsertClient(c); err == nil {
			s.dirtyClients.Delete(id)
		}
		return true
	})
}

var globalLock sync.Mutex

// FlushGlobalToStore is the Glob singleton equivalent of FlushHostsToStore.
func (s *JsonDb) FlushGlobalToStore() {
	globalLock.Lock()
	defer globalLock.Unlock()
	if s.GlobalStore == nil || s.Global == nil {
		return
	}
	if !s.dirtyGlobal.Load() {
		return
	}
	if err := s.GlobalStore.UpsertGlobal(s.Global); err == nil {
		s.dirtyGlobal.Store(false)
	}
}

func (s *JsonDb) GetClientId() int32 {
	return atomic.AddInt32(&s.ClientIncreaseId, 1)
}

func (s *JsonDb) GetTaskId() int32 {
	return atomic.AddInt32(&s.TaskIncreaseId, 1)
}

func (s *JsonDb) GetHostId() int32 {
	return atomic.AddInt32(&s.HostIncreaseId, 1)
}

// loadSyncMapFromFile reads a legacy newline-delimited JSON snapshot
// (conf/clients.json, conf/tasks.json, conf/hosts.json, conf/api_tokens.json).
// Phase 7: these files are no longer written and may not exist; missing
// or empty files are silently ignored — SQLite is the source of truth
// at runtime, so an empty in-memory map at startup just means SQLite
// will populate it via BackfillOrLoad*.
func loadSyncMapFromFile(filePath string, f func(value string)) {
	if !common.FileExists(filePath) {
		return
	}
	b, err := common.ReadAllFromFile(filePath)
	if err != nil {
		return
	}
	for _, v := range strings.Split(string(b), "\n"+common.CONN_DATA_SEQ) {
		f(v)
	}
}

// loadSyncMapFromFileWithSingleJson reads a legacy single-document JSON
// file (conf/global.json). Same Phase-7 semantics as loadSyncMapFromFile:
// missing or unreadable files are ignored.
func loadSyncMapFromFileWithSingleJson(filePath string, f func(value string)) {
	if !common.FileExists(filePath) {
		return
	}
	b, err := common.ReadAllFromFile(filePath)
	if err != nil {
		return
	}
	f(string(b))
}
