/**
 * nps-table.js — Tailwind-styled drop-in replacement for bootstrap-table.
 * Implements the subset of bootstrapTable API used by NPS admin panel.
 * Depends on: jQuery (already loaded), FontAwesome (already loaded).
 */
(function ($) {
    'use strict';

    /* ------------------------------------------------------------------ */
    /* Constructor                                                          */
    /* ------------------------------------------------------------------ */
    function NpsTable(element, options) {
        this.$el      = $(element);
        this.opts     = $.extend(true, {
            method          : 'post',
            contentType     : 'application/x-www-form-urlencoded',
            striped         : true,
            search          : true,
            showRefresh     : true,
            pagination      : true,
            sidePagination  : 'server',
            pageNumber      : 1,
            pageList        : [10, 25, 50, 100],
            detailView      : false,
            columns         : []
        }, options);

        this.page          = this.opts.pageNumber || 1;
        this.pageSize      = (this.opts.pageList && this.opts.pageList[0]) || 10;
        this.searchKeyword = '';
        this.sortField     = '';
        this.sortOrder     = '';
        this.expandedRows  = {};
        this.locale        = null;

        this._build();
        this._bindEvents();
        this.load();
    }

    /* ------------------------------------------------------------------ */
    /* Build DOM                                                            */
    /* ------------------------------------------------------------------ */
    NpsTable.prototype._build = function () {
        var self = this;

        var $wrap = $('<div class="nps-table-wrap space-y-3">');

        /* Top bar */
        var $topbar = $('<div class="flex flex-wrap items-center justify-between gap-2">');
        self.$toolbarSlot = $('<div class="nps-toolbar-slot flex items-center gap-2 flex-wrap">');
        var $right = $('<div class="flex items-center gap-2">');

        if (self.opts.search) {
            self.$searchBox = $('<input type="text" placeholder="Search…"'
                + ' class="px-3 py-1.5 text-sm border border-gray-300 rounded-lg'
                + ' focus:outline-none focus:ring-2 focus:ring-indigo-400 w-44 bg-white">');
            $right.append(self.$searchBox);
        }
        if (self.opts.showRefresh) {
            self.$refreshBtn = $('<button title="Refresh"'
                + ' class="px-2.5 py-1.5 text-sm bg-white border border-gray-300 rounded-lg'
                + ' hover:bg-gray-50 text-gray-500 transition-colors">'
                + '<i class="fa fa-sync-alt"></i></button>');
            $right.append(self.$refreshBtn);
        }

        $topbar.append(self.$toolbarSlot).append($right);

        /* Table */
        var $tableWrap = $('<div class="overflow-x-auto rounded-xl border border-gray-200 shadow-sm bg-white">');
        var $table = $('<table class="w-full text-sm text-gray-700 border-collapse">');

        /* thead */
        self.$thead = $('<thead class="text-xs font-semibold text-gray-500 uppercase bg-gray-50 border-b border-gray-200">');
        var $theadTr = $('<tr>');
        if (self.opts.detailView) {
            $theadTr.append('<th class="px-3 py-3 w-8"></th>');
        }
        $.each(self.opts.columns, function (i, col) {
            if (col.visible === false) return;
            var $th = $('<th class="px-4 py-3 whitespace-nowrap text-left">');
            if (col.halign === 'center' || col.align === 'center') $th.addClass('text-center');
            $th.html(col.title || col.field || '');
            if (col.sortable) {
                $th.addClass('cursor-pointer select-none nps-sortable');
                $th.attr('data-field', col.field);
                $th.append(' <i class="fa fa-sort text-gray-300 text-xs ml-0.5"></i>');
            }
            $theadTr.append($th);
        });
        self.$thead.append($theadTr);

        /* tbody */
        self.$tbody = $('<tbody class="divide-y divide-gray-100">');

        $table.append(self.$thead).append(self.$tbody);
        $tableWrap.append($table);

        /* Pagination */
        self.$pager = $('<div class="flex flex-wrap items-center justify-between gap-3 text-sm text-gray-600 px-1">');

        $wrap.append($topbar).append($tableWrap).append(self.$pager);
        self.$el.empty().append($wrap);

        /* Move toolbar into slot */
        if (self.opts.toolbar) {
            var $tb = $(self.opts.toolbar);
            if ($tb.length) {
                self.$toolbarSlot.append($tb.children().clone(true));
                $tb.hide();
            }
        }
    };

    /* ------------------------------------------------------------------ */
    /* Events                                                               */
    /* ------------------------------------------------------------------ */
    NpsTable.prototype._bindEvents = function () {
        var self = this;

        if (self.$searchBox) {
            self.$searchBox.on('input', function () {
                self.searchKeyword = $(this).val();
                self.page = 1;
                clearTimeout(self._searchTimer);
                self._searchTimer = setTimeout(function () { self.load(); }, 320);
            });
        }

        if (self.$refreshBtn) {
            self.$refreshBtn.on('click', function () { self.load(); });
        }

        self.$thead.on('click', '.nps-sortable', function () {
            var field = $(this).data('field');
            if (self.sortField === field) {
                self.sortOrder = self.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                self.sortField = field;
                self.sortOrder = 'asc';
            }
            self.$thead.find('.nps-sortable i')
                .removeClass('fa-sort-up fa-sort-down text-indigo-500')
                .addClass('fa-sort text-gray-300');
            $(this).find('i')
                .removeClass('fa-sort text-gray-300')
                .addClass(self.sortOrder === 'asc' ? 'fa-sort-up text-indigo-500' : 'fa-sort-down text-indigo-500');
            self.load();
        });
    };

    /* ------------------------------------------------------------------ */
    /* Load data                                                            */
    /* ------------------------------------------------------------------ */
    NpsTable.prototype.load = function () {
        var self = this;

        var params = {
            offset : (self.page - 1) * self.pageSize,
            limit  : self.pageSize,
            search : self.searchKeyword,
            sort   : self.sortField,
            order  : self.sortOrder
        };

        if (typeof self.opts.queryParams === 'function') {
            params = self.opts.queryParams(params);
        }

        self.$tbody.html(
            '<tr><td colspan="99" class="px-4 py-10 text-center text-gray-400">'
            + '<i class="fa fa-circle-notch fa-spin mr-2 text-indigo-400"></i>Loading…</td></tr>'
        );

        $.ajax({
            type        : self.opts.method,
            url         : self.opts.url,
            data        : params,
            contentType : self.opts.contentType,
            success: function (data) {
                self.totalRows = data.total || 0;
                self._renderRows(data.rows || []);
                self._renderPager();
                if (typeof self.opts.onLoadSuccess === 'function') {
                    self.opts.onLoadSuccess.call(self.$el[0], data);
                }
                if (typeof self.opts.onPostBody === 'function') {
                    self.opts.onPostBody.call(self.$el[0], data);
                }
                if (typeof $('body').setLang === 'function') {
                    $('body').setLang('');
                }
            },
            error: function () {
                self.$tbody.html(
                    '<tr><td colspan="99" class="px-4 py-6 text-center text-red-500">'
                    + '<i class="fa fa-exclamation-circle mr-1"></i>Load failed</td></tr>'
                );
            }
        });
    };

    /* ------------------------------------------------------------------ */
    /* Render rows                                                          */
    /* ------------------------------------------------------------------ */
    NpsTable.prototype._renderRows = function (rows) {
        var self = this;
        self.$tbody.empty();
        self.expandedRows = {};

        var visibleCols = $.grep(self.opts.columns, function (c) { return c.visible !== false; });

        if (!rows || rows.length === 0) {
            self.$tbody.html(
                '<tr><td colspan="99" class="px-4 py-10 text-center text-gray-400">'
                + '<i class="fa fa-inbox mr-2"></i>No data</td></tr>'
            );
            return;
        }

        $.each(rows, function (idx, row) {
            var $tr = $('<tr class="hover:bg-blue-50/30 transition-colors">');
            if (self.opts.striped && idx % 2 === 1) $tr.addClass('bg-gray-50/60');

            if (self.opts.detailView) {
                var $expandTd = $('<td class="px-3 py-3 text-center">');
                var $btn = $(
                    '<button data-row="' + idx + '"'
                    + ' class="nps-expand w-5 h-5 rounded border border-gray-300 text-gray-500'
                    + ' hover:bg-indigo-50 hover:border-indigo-400 hover:text-indigo-600'
                    + ' text-xs inline-flex items-center justify-center transition-colors">'
                    + '<i class="fa fa-plus"></i></button>'
                );
                $expandTd.append($btn);
                $tr.append($expandTd);
            }

            $.each(visibleCols, function (ci, col) {
                var val  = self._getVal(row, col.field);
                var html = (typeof col.formatter === 'function')
                    ? col.formatter(val, row, idx)
                    : (val !== undefined && val !== null ? String(val) : '');
                var $td = $('<td class="px-4 py-3">');
                if (col.align === 'center' || col.halign === 'center') $td.addClass('text-center');
                $td.html(html);
                $tr.append($td);
            });

            self.$tbody.append($tr);

            /* Expand row */
            $tr.find('.nps-expand').on('click', function () {
                var ri = parseInt($(this).data('row'));
                if (self.expandedRows[ri]) {
                    self.$tbody.find('.nps-detail-row[data-p="' + ri + '"]').remove();
                    delete self.expandedRows[ri];
                    $(this).find('i').removeClass('fa-minus').addClass('fa-plus');
                    $(this).removeClass('bg-indigo-50 border-indigo-400 text-indigo-600')
                           .addClass('border-gray-300 text-gray-500');
                } else {
                    var detailHtml = '';
                    if (typeof self.opts.detailFormatter === 'function') {
                        detailHtml = self.opts.detailFormatter(ri, row, $tr);
                    }
                    var span = visibleCols.length + 1;
                    var $dtr = $(
                        '<tr class="nps-detail-row detail-view bg-indigo-50/40" data-p="' + ri + '">'
                        + '<td></td>'
                        + '<td colspan="' + span + '" class="px-5 py-3 text-xs text-gray-600 leading-relaxed">'
                        + detailHtml + '</td></tr>'
                    );
                    $tr.after($dtr);
                    self.expandedRows[ri] = true;
                    $(this).find('i').removeClass('fa-plus').addClass('fa-minus');
                    $(this).addClass('bg-indigo-50 border-indigo-400 text-indigo-600')
                           .removeClass('border-gray-300 text-gray-500');
                    if (typeof self.opts.onExpandRow === 'function') {
                        self.opts.onExpandRow.call(self.$el[0], ri, row, $dtr);
                    }
                    if (typeof $('body').setLang === 'function') {
                        $('body').setLang('.detail-view');
                    }
                }
            });
        });
    };

    /* ------------------------------------------------------------------ */
    /* Render pagination                                                    */
    /* ------------------------------------------------------------------ */
    NpsTable.prototype._renderPager = function () {
        var self = this;
        self.$pager.empty();
        if (!self.opts.pagination || self.totalRows === 0) return;

        var totalPages = Math.max(1, Math.ceil(self.totalRows / self.pageSize));
        var start = (self.page - 1) * self.pageSize + 1;
        var end   = Math.min(self.page * self.pageSize, self.totalRows);

        /* Left: info + page size */
        var $left = $('<div class="flex items-center gap-3">');
        $left.append('<span class="text-gray-500 text-xs">' + start + '–' + end + ' / ' + self.totalRows + '</span>');

        var $sel = $('<select class="text-xs border border-gray-300 rounded-lg px-2 py-1'
            + ' focus:outline-none focus:ring-1 focus:ring-indigo-400 bg-white">');
        $.each(self.opts.pageList, function (i, s) {
            $sel.append($('<option>').val(s).text(s + ' / page').prop('selected', s === self.pageSize));
        });
        $sel.on('change', function () {
            self.pageSize = parseInt($(this).val());
            self.page = 1;
            self.load();
        });
        $left.append($sel);

        /* Right: page buttons */
        var $right = $('<div class="flex items-center gap-1">');
        var base    = 'inline-flex items-center justify-center min-w-8 h-8 px-2 rounded-lg text-xs border transition-colors ';
        var active  = base + 'bg-indigo-600 text-white border-indigo-600 font-semibold';
        var normal  = base + 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50 cursor-pointer';
        var disabled = base + 'bg-white text-gray-300 border-gray-200 cursor-not-allowed';

        var $prev = $('<button class="' + (self.page <= 1 ? disabled : normal) + '">'
            + '<i class="fa fa-chevron-left text-xs"></i></button>');
        if (self.page > 1) $prev.on('click', function () { self.page--; self.load(); });
        $right.append($prev);

        var pStart = Math.max(1, self.page - 2);
        var pEnd   = Math.min(totalPages, pStart + 4);
        for (var p = pStart; p <= pEnd; p++) {
            (function (pg) {
                var $btn = $('<button class="' + (pg === self.page ? active : normal) + '">' + pg + '</button>');
                if (pg !== self.page) $btn.on('click', function () { self.page = pg; self.load(); });
                $right.append($btn);
            })(p);
        }

        var $next = $('<button class="' + (self.page >= totalPages ? disabled : normal) + '">'
            + '<i class="fa fa-chevron-right text-xs"></i></button>');
        if (self.page < totalPages) $next.on('click', function () { self.page++; self.load(); });
        $right.append($next);

        self.$pager.append($left).append($right);
    };

    /* ------------------------------------------------------------------ */
    /* Helpers                                                              */
    /* ------------------------------------------------------------------ */
    NpsTable.prototype._getVal = function (obj, field) {
        if (!field) return '';
        var parts = field.split('.');
        var v = obj;
        for (var i = 0; i < parts.length; i++) {
            if (v === undefined || v === null) return '';
            v = v[parts[i]];
        }
        return v;
    };

    /* ------------------------------------------------------------------ */
    /* jQuery plugin                                                        */
    /* ------------------------------------------------------------------ */
    $.fn.bootstrapTable = function (options) {
        var args = Array.prototype.slice.call(arguments, 1);

        if (typeof options === 'string') {
            return this.each(function () {
                var inst = $(this).data('npsTable');
                if (!inst) return;
                switch (options) {
                    case 'refresh':
                        inst.load();
                        break;
                    case 'refreshOptions':
                        $.extend(inst.opts, args[0] || {});
                        break;
                    case 'getData':
                        break;
                }
            });
        }

        return this.each(function () {
            var inst = new NpsTable(this, options);
            $(this).data('npsTable', inst);
        });
    };

    /* ------------------------------------------------------------------ */
    /* Global helpers used by NPS templates                                 */
    /* ------------------------------------------------------------------ */
    window.encodeToBase64 = function (str) {
        try { return btoa(unescape(encodeURIComponent(str))); } catch (e) { return str; }
    };

    window.copyCommand = function (btn) {
        var text = $(btn).prev('span').text() || $(btn).data('clipboard-text');
        if (!text) return;
        try {
            navigator.clipboard.writeText(text).catch(function () {
                _fallbackCopy(text);
            });
        } catch (e) { _fallbackCopy(text); }
    };

    function _fallbackCopy(text) {
        var ta = document.createElement('textarea');
        ta.value = text;
        ta.style.cssText = 'position:fixed;opacity:0';
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
    }

})(jQuery);
