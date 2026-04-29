import{d as T,l as t,I as te,E as S,a5 as re,a6 as ie,a7 as ne,a8 as se,a9 as oe,H as O,v as h,x as P,y as K,z as M,D as U,aa as ae,ab as Y,ac as le,ad as ce,ae as ue,X as de,k as E,af as fe,ag as pe,ah as ge,G as V,ai as q,aj as he,ak as ve,al as me,u as ye,a as be,am as xe,an as Ce,c as J,e as n,a0 as Z,w as v,T as $e,b as d,ao as Se,o as G,j as k,t as x,N as R,a1 as $}from"./index-CHE8zU3A.js";import{N as Q}from"./Alert-Bvh1lF2g.js";import{u as we,N as D}from"./Tag-DSYe8uvd.js";import{N as ze,a as I}from"./Grid-CWDfzZFE.js";import{f as j}from"./format-length-B-p6aW7q.js";import{N as F}from"./Space-DBW4KciJ.js";import"./next-frame-once-C5Ksf8W7.js";const Pe={success:t(se,null),error:t(ne,null),warning:t(ie,null),info:t(re,null)},ke=T({name:"ProgressCircle",props:{clsPrefix:{type:String,required:!0},status:{type:String,required:!0},strokeWidth:{type:Number,required:!0},fillColor:[String,Object],railColor:String,railStyle:[String,Object],percentage:{type:Number,default:0},offsetDegree:{type:Number,default:0},showIndicator:{type:Boolean,required:!0},indicatorTextColor:String,unit:String,viewBoxWidth:{type:Number,required:!0},gapDegree:{type:Number,required:!0},gapOffsetDegree:{type:Number,default:0}},setup(e,{slots:s}){const l=S(()=>{const o="gradient",{fillColor:i}=e;return typeof i=="object"?`${o}-${oe(JSON.stringify(i))}`:o});function r(o,i,a,u){const{gapDegree:b,viewBoxWidth:c,strokeWidth:m}=e,g=50,C=0,y=g,f=0,z=2*g,_=50+m/2,w=`M ${_},${_} m ${C},${y}
      a ${g},${g} 0 1 1 ${f},${-z}
      a ${g},${g} 0 1 1 ${-f},${z}`,B=Math.PI*2*g,N={stroke:u==="rail"?a:typeof e.fillColor=="object"?`url(#${l.value})`:a,strokeDasharray:`${Math.min(o,100)/100*(B-b)}px ${c*8}px`,strokeDashoffset:`-${b/2}px`,transformOrigin:i?"center":void 0,transform:i?`rotate(${i}deg)`:void 0};return{pathString:w,pathStyle:N}}const p=()=>{const o=typeof e.fillColor=="object",i=o?e.fillColor.stops[0]:"",a=o?e.fillColor.stops[1]:"";return o&&t("defs",null,t("linearGradient",{id:l.value,x1:"0%",y1:"100%",x2:"100%",y2:"0%"},t("stop",{offset:"0%","stop-color":i}),t("stop",{offset:"100%","stop-color":a})))};return()=>{const{fillColor:o,railColor:i,strokeWidth:a,offsetDegree:u,status:b,percentage:c,showIndicator:m,indicatorTextColor:g,unit:C,gapOffsetDegree:y,clsPrefix:f}=e,{pathString:z,pathStyle:_}=r(100,0,i,"rail"),{pathString:w,pathStyle:B}=r(c,u,o,"fill"),N=100+a;return t("div",{class:`${f}-progress-content`,role:"none"},t("div",{class:`${f}-progress-graph`,"aria-hidden":!0},t("div",{class:`${f}-progress-graph-circle`,style:{transform:y?`rotate(${y}deg)`:void 0}},t("svg",{viewBox:`0 0 ${N} ${N}`},p(),t("g",null,t("path",{class:`${f}-progress-graph-circle-rail`,d:z,"stroke-width":a,"stroke-linecap":"round",fill:"none",style:_})),t("g",null,t("path",{class:[`${f}-progress-graph-circle-fill`,c===0&&`${f}-progress-graph-circle-fill--empty`],d:w,"stroke-width":a,"stroke-linecap":"round",fill:"none",style:B}))))),m?t("div",null,s.default?t("div",{class:`${f}-progress-custom-content`,role:"none"},s.default()):b!=="default"?t("div",{class:`${f}-progress-icon`,"aria-hidden":!0},t(te,{clsPrefix:f},{default:()=>Pe[b]})):t("div",{class:`${f}-progress-text`,style:{color:g},role:"none"},t("span",{class:`${f}-progress-text__percentage`},c),t("span",{class:`${f}-progress-text__unit`},C))):null)}}}),_e={success:t(se,null),error:t(ne,null),warning:t(ie,null),info:t(re,null)},Be=T({name:"ProgressLine",props:{clsPrefix:{type:String,required:!0},percentage:{type:Number,default:0},railColor:String,railStyle:[String,Object],fillColor:[String,Object],status:{type:String,required:!0},indicatorPlacement:{type:String,required:!0},indicatorTextColor:String,unit:{type:String,default:"%"},processing:{type:Boolean,required:!0},showIndicator:{type:Boolean,required:!0},height:[String,Number],railBorderRadius:[String,Number],fillBorderRadius:[String,Number]},setup(e,{slots:s}){const l=S(()=>j(e.height)),r=S(()=>{var i,a;return typeof e.fillColor=="object"?`linear-gradient(to right, ${(i=e.fillColor)===null||i===void 0?void 0:i.stops[0]} , ${(a=e.fillColor)===null||a===void 0?void 0:a.stops[1]})`:e.fillColor}),p=S(()=>e.railBorderRadius!==void 0?j(e.railBorderRadius):e.height!==void 0?j(e.height,{c:.5}):""),o=S(()=>e.fillBorderRadius!==void 0?j(e.fillBorderRadius):e.railBorderRadius!==void 0?j(e.railBorderRadius):e.height!==void 0?j(e.height,{c:.5}):"");return()=>{const{indicatorPlacement:i,railColor:a,railStyle:u,percentage:b,unit:c,indicatorTextColor:m,status:g,showIndicator:C,processing:y,clsPrefix:f}=e;return t("div",{class:`${f}-progress-content`,role:"none"},t("div",{class:`${f}-progress-graph`,"aria-hidden":!0},t("div",{class:[`${f}-progress-graph-line`,{[`${f}-progress-graph-line--indicator-${i}`]:!0}]},t("div",{class:`${f}-progress-graph-line-rail`,style:[{backgroundColor:a,height:l.value,borderRadius:p.value},u]},t("div",{class:[`${f}-progress-graph-line-fill`,y&&`${f}-progress-graph-line-fill--processing`],style:{maxWidth:`${e.percentage}%`,background:r.value,height:l.value,lineHeight:l.value,borderRadius:o.value}},i==="inside"?t("div",{class:`${f}-progress-graph-line-indicator`,style:{color:m}},s.default?s.default():`${b}${c}`):null)))),C&&i==="outside"?t("div",null,s.default?t("div",{class:`${f}-progress-custom-content`,style:{color:m},role:"none"},s.default()):g==="default"?t("div",{role:"none",class:`${f}-progress-icon ${f}-progress-icon--as-text`,style:{color:m}},b,c):t("div",{class:`${f}-progress-icon`,"aria-hidden":!0},t(te,{clsPrefix:f},{default:()=>_e[g]}))):null)}}});function ee(e,s,l=100){return`m ${l/2} ${l/2-e} a ${e} ${e} 0 1 1 0 ${2*e} a ${e} ${e} 0 1 1 0 -${2*e}`}const Ne=T({name:"ProgressMultipleCircle",props:{clsPrefix:{type:String,required:!0},viewBoxWidth:{type:Number,required:!0},percentage:{type:Array,default:[0]},strokeWidth:{type:Number,required:!0},circleGap:{type:Number,required:!0},showIndicator:{type:Boolean,required:!0},fillColor:{type:Array,default:()=>[]},railColor:{type:Array,default:()=>[]},railStyle:{type:Array,default:()=>[]}},setup(e,{slots:s}){const l=S(()=>e.percentage.map((o,i)=>`${Math.PI*o/100*(e.viewBoxWidth/2-e.strokeWidth/2*(1+2*i)-e.circleGap*i)*2}, ${e.viewBoxWidth*8}`)),r=(p,o)=>{const i=e.fillColor[o],a=typeof i=="object"?i.stops[0]:"",u=typeof i=="object"?i.stops[1]:"";return typeof e.fillColor[o]=="object"&&t("linearGradient",{id:`gradient-${o}`,x1:"100%",y1:"0%",x2:"0%",y2:"100%"},t("stop",{offset:"0%","stop-color":a}),t("stop",{offset:"100%","stop-color":u}))};return()=>{const{viewBoxWidth:p,strokeWidth:o,circleGap:i,showIndicator:a,fillColor:u,railColor:b,railStyle:c,percentage:m,clsPrefix:g}=e;return t("div",{class:`${g}-progress-content`,role:"none"},t("div",{class:`${g}-progress-graph`,"aria-hidden":!0},t("div",{class:`${g}-progress-graph-circle`},t("svg",{viewBox:`0 0 ${p} ${p}`},t("defs",null,m.map((C,y)=>r(C,y))),m.map((C,y)=>t("g",{key:y},t("path",{class:`${g}-progress-graph-circle-rail`,d:ee(p/2-o/2*(1+2*y)-i*y,o,p),"stroke-width":o,"stroke-linecap":"round",fill:"none",style:[{strokeDashoffset:0,stroke:b[y]},c[y]]}),t("path",{class:[`${g}-progress-graph-circle-fill`,C===0&&`${g}-progress-graph-circle-fill--empty`],d:ee(p/2-o/2*(1+2*y)-i*y,o,p),"stroke-width":o,"stroke-linecap":"round",fill:"none",style:{strokeDasharray:l.value[y],strokeDashoffset:0,stroke:typeof u[y]=="object"?`url(#gradient-${y})`:u[y]}})))))),a&&s.default?t("div",null,t("div",{class:`${g}-progress-text`},s.default())):null)}}}),Re=O([h("progress",{display:"inline-block"},[h("progress-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 `),P("line",`
 width: 100%;
 display: block;
 `,[h("progress-content",`
 display: flex;
 align-items: center;
 `,[h("progress-graph",{flex:1})]),h("progress-custom-content",{marginLeft:"14px"}),h("progress-icon",`
 width: 30px;
 padding-left: 14px;
 height: var(--n-icon-size-line);
 line-height: var(--n-icon-size-line);
 font-size: var(--n-icon-size-line);
 `,[P("as-text",`
 color: var(--n-text-color-line-outer);
 text-align: center;
 width: 40px;
 font-size: var(--n-font-size);
 padding-left: 4px;
 transition: color .3s var(--n-bezier);
 `)])]),P("circle, dashboard",{width:"120px"},[h("progress-custom-content",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 justify-content: center;
 `),h("progress-text",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 color: inherit;
 font-size: var(--n-font-size-circle);
 color: var(--n-text-color-circle);
 font-weight: var(--n-font-weight-circle);
 transition: color .3s var(--n-bezier);
 white-space: nowrap;
 `),h("progress-icon",`
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 color: var(--n-icon-color);
 font-size: var(--n-icon-size-circle);
 `)]),P("multiple-circle",`
 width: 200px;
 color: inherit;
 `,[h("progress-text",`
 font-weight: var(--n-font-weight-circle);
 color: var(--n-text-color-circle);
 position: absolute;
 left: 50%;
 top: 50%;
 transform: translateX(-50%) translateY(-50%);
 display: flex;
 align-items: center;
 justify-content: center;
 transition: color .3s var(--n-bezier);
 `)]),h("progress-content",{position:"relative"}),h("progress-graph",{position:"relative"},[h("progress-graph-circle",[O("svg",{verticalAlign:"bottom"}),h("progress-graph-circle-fill",`
 stroke: var(--n-fill-color);
 transition:
 opacity .3s var(--n-bezier),
 stroke .3s var(--n-bezier),
 stroke-dasharray .3s var(--n-bezier);
 `,[P("empty",{opacity:0})]),h("progress-graph-circle-rail",`
 transition: stroke .3s var(--n-bezier);
 overflow: hidden;
 stroke: var(--n-rail-color);
 `)]),h("progress-graph-line",[P("indicator-inside",[h("progress-graph-line-rail",`
 height: 16px;
 line-height: 16px;
 border-radius: 10px;
 `,[h("progress-graph-line-fill",`
 height: inherit;
 border-radius: 10px;
 `),h("progress-graph-line-indicator",`
 background: #0000;
 white-space: nowrap;
 text-align: right;
 margin-left: 14px;
 margin-right: 14px;
 height: inherit;
 font-size: 12px;
 color: var(--n-text-color-line-inner);
 transition: color .3s var(--n-bezier);
 `)])]),P("indicator-inside-label",`
 height: 16px;
 display: flex;
 align-items: center;
 `,[h("progress-graph-line-rail",`
 flex: 1;
 transition: background-color .3s var(--n-bezier);
 `),h("progress-graph-line-indicator",`
 background: var(--n-fill-color);
 font-size: 12px;
 transform: translateZ(0);
 display: flex;
 vertical-align: middle;
 height: 16px;
 line-height: 16px;
 padding: 0 10px;
 border-radius: 10px;
 position: absolute;
 white-space: nowrap;
 color: var(--n-text-color-line-inner);
 transition:
 right .2s var(--n-bezier),
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 `)]),h("progress-graph-line-rail",`
 position: relative;
 overflow: hidden;
 height: var(--n-rail-height);
 border-radius: 5px;
 background-color: var(--n-rail-color);
 transition: background-color .3s var(--n-bezier);
 `,[h("progress-graph-line-fill",`
 background: var(--n-fill-color);
 position: relative;
 border-radius: 5px;
 height: inherit;
 width: 100%;
 max-width: 0%;
 transition:
 background-color .3s var(--n-bezier),
 max-width .2s var(--n-bezier);
 `,[P("processing",[O("&::after",`
 content: "";
 background-image: var(--n-line-bg-processing);
 animation: progress-processing-animation 2s var(--n-bezier) infinite;
 `)])])])])])]),O("@keyframes progress-processing-animation",`
 0% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 100%;
 opacity: 1;
 }
 66% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 0;
 opacity: 0;
 }
 100% {
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 right: 0;
 opacity: 0;
 }
 `)]),Ie=Object.assign(Object.assign({},M.props),{processing:Boolean,type:{type:String,default:"line"},gapDegree:Number,gapOffsetDegree:Number,status:{type:String,default:"default"},railColor:[String,Array],railStyle:[String,Array],color:[String,Array,Object],viewBoxWidth:{type:Number,default:100},strokeWidth:{type:Number,default:7},percentage:[Number,Array],unit:{type:String,default:"%"},showIndicator:{type:Boolean,default:!0},indicatorPosition:{type:String,default:"outside"},indicatorPlacement:{type:String,default:"outside"},indicatorTextColor:String,circleGap:{type:Number,default:1},height:Number,borderRadius:[String,Number],fillBorderRadius:[String,Number],offsetDegree:Number}),X=T({name:"Progress",props:Ie,setup(e){const s=S(()=>e.indicatorPlacement||e.indicatorPosition),l=S(()=>{if(e.gapDegree||e.gapDegree===0)return e.gapDegree;if(e.type==="dashboard")return 75}),{mergedClsPrefixRef:r,inlineThemeDisabled:p}=K(e),o=M("Progress","-progress",Re,ae,e,r),i=S(()=>{const{status:u}=e,{common:{cubicBezierEaseInOut:b},self:{fontSize:c,fontSizeCircle:m,railColor:g,railHeight:C,iconSizeCircle:y,iconSizeLine:f,textColorCircle:z,textColorLineInner:_,textColorLineOuter:w,lineBgProcessing:B,fontWeightCircle:N,[Y("iconColor",u)]:L,[Y("fillColor",u)]:A}}=o.value;return{"--n-bezier":b,"--n-fill-color":A,"--n-font-size":c,"--n-font-size-circle":m,"--n-font-weight-circle":N,"--n-icon-color":L,"--n-icon-size-circle":y,"--n-icon-size-line":f,"--n-line-bg-processing":B,"--n-rail-color":g,"--n-rail-height":C,"--n-text-color-circle":z,"--n-text-color-line-inner":_,"--n-text-color-line-outer":w}}),a=p?U("progress",S(()=>e.status[0]),i,e):void 0;return{mergedClsPrefix:r,mergedIndicatorPlacement:s,gapDeg:l,cssVars:p?void 0:i,themeClass:a?.themeClass,onRender:a?.onRender}},render(){const{type:e,cssVars:s,indicatorTextColor:l,showIndicator:r,status:p,railColor:o,railStyle:i,color:a,percentage:u,viewBoxWidth:b,strokeWidth:c,mergedIndicatorPlacement:m,unit:g,borderRadius:C,fillBorderRadius:y,height:f,processing:z,circleGap:_,mergedClsPrefix:w,gapDeg:B,gapOffsetDegree:N,themeClass:L,$slots:A,onRender:H}=this;return H?.(),t("div",{class:[L,`${w}-progress`,`${w}-progress--${e}`,`${w}-progress--${p}`],style:s,"aria-valuemax":100,"aria-valuemin":0,"aria-valuenow":u,role:e==="circle"||e==="line"||e==="dashboard"?"progressbar":"none"},e==="circle"||e==="dashboard"?t(ke,{clsPrefix:w,status:p,showIndicator:r,indicatorTextColor:l,railColor:o,fillColor:a,railStyle:i,offsetDegree:this.offsetDegree,percentage:u,viewBoxWidth:b,strokeWidth:c,gapDegree:B===void 0?e==="dashboard"?75:0:B,gapOffsetDegree:N,unit:g},A):e==="line"?t(Be,{clsPrefix:w,status:p,showIndicator:r,indicatorTextColor:l,railColor:o,fillColor:a,railStyle:i,percentage:u,processing:z,indicatorPlacement:m,unit:g,fillBorderRadius:y,railBorderRadius:C,height:f},A):e==="multiple-circle"?t(Ne,{clsPrefix:w,strokeWidth:c,railColor:o,fillColor:a,railStyle:i,viewBoxWidth:b,percentage:u,showIndicator:r,circleGap:_},A):null)}}),Te=O([O("@keyframes spin-rotate",`
 from {
 transform: rotate(0);
 }
 to {
 transform: rotate(360deg);
 }
 `),h("spin-container",`
 position: relative;
 `,[h("spin-body",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 `,[le()])]),h("spin-body",`
 display: inline-flex;
 align-items: center;
 justify-content: center;
 flex-direction: column;
 `),h("spin",`
 display: inline-flex;
 height: var(--n-size);
 width: var(--n-size);
 font-size: var(--n-size);
 color: var(--n-color);
 `,[P("rotate",`
 animation: spin-rotate 2s linear infinite;
 `)]),h("spin-description",`
 display: inline-block;
 font-size: var(--n-font-size);
 color: var(--n-text-color);
 transition: color .3s var(--n-bezier);
 margin-top: 8px;
 `),h("spin-content",`
 opacity: 1;
 transition: opacity .3s var(--n-bezier);
 pointer-events: all;
 `,[P("spinning",`
 user-select: none;
 -webkit-user-select: none;
 pointer-events: none;
 opacity: var(--n-opacity-spinning);
 `)])]),De={small:20,medium:18,large:16},je=Object.assign(Object.assign(Object.assign({},M.props),{contentClass:String,contentStyle:[Object,String],description:String,size:{type:[String,Number],default:"medium"},show:{type:Boolean,default:!0},rotate:{type:Boolean,default:!0},spinning:{type:Boolean,validator:()=>!0,default:void 0},delay:Number}),fe),We=T({name:"Spin",props:je,slots:Object,setup(e){const{mergedClsPrefixRef:s,inlineThemeDisabled:l}=K(e),r=M("Spin","-spin",Te,pe,e,s),p=S(()=>{const{size:u}=e,{common:{cubicBezierEaseInOut:b},self:c}=r.value,{opacitySpinning:m,color:g,textColor:C}=c,y=typeof u=="number"?ge(u):c[Y("size",u)];return{"--n-bezier":b,"--n-opacity-spinning":m,"--n-size":y,"--n-color":g,"--n-text-color":C}}),o=l?U("spin",S(()=>{const{size:u}=e;return typeof u=="number"?String(u):u[0]}),p,e):void 0,i=we(e,["spinning","show"]),a=E(!1);return de(u=>{let b;if(i.value){const{delay:c}=e;if(c){b=window.setTimeout(()=>{a.value=!0},c),u(()=>{clearTimeout(b)});return}}a.value=i.value}),{mergedClsPrefix:s,active:a,mergedStrokeWidth:S(()=>{const{strokeWidth:u}=e;if(u!==void 0)return u;const{size:b}=e;return De[typeof b=="number"?"medium":b]}),cssVars:l?void 0:p,themeClass:o?.themeClass,onRender:o?.onRender}},render(){var e,s;const{$slots:l,mergedClsPrefix:r,description:p}=this,o=l.icon&&this.rotate,i=(p||l.description)&&t("div",{class:`${r}-spin-description`},p||((e=l.description)===null||e===void 0?void 0:e.call(l))),a=l.icon?t("div",{class:[`${r}-spin-body`,this.themeClass]},t("div",{class:[`${r}-spin`,o&&`${r}-spin--rotate`],style:l.default?"":this.cssVars},l.icon()),i):t("div",{class:[`${r}-spin-body`,this.themeClass]},t(ce,{clsPrefix:r,style:l.default?"":this.cssVars,stroke:this.stroke,"stroke-width":this.mergedStrokeWidth,radius:this.radius,scale:this.scale,class:`${r}-spin`}),i);return(s=this.onRender)===null||s===void 0||s.call(this),l.default?t("div",{class:[`${r}-spin-container`,this.themeClass],style:this.cssVars},t("div",{class:[`${r}-spin-content`,this.active&&`${r}-spin-content--spinning`,this.contentClass],style:this.contentStyle},l),t(ue,{name:"fade-in-transition"},{default:()=>this.active?a:null})):a}}),Oe=h("statistic",[V("label",`
 font-weight: var(--n-label-font-weight);
 transition: .3s color var(--n-bezier);
 font-size: var(--n-label-font-size);
 color: var(--n-label-text-color);
 `),h("statistic-value",`
 margin-top: 4px;
 font-weight: var(--n-value-font-weight);
 `,[V("prefix",`
 margin: 0 4px 0 0;
 font-size: var(--n-value-font-size);
 transition: .3s color var(--n-bezier);
 color: var(--n-value-prefix-text-color);
 `,[h("icon",{verticalAlign:"-0.125em"})]),V("content",`
 font-size: var(--n-value-font-size);
 transition: .3s color var(--n-bezier);
 color: var(--n-value-text-color);
 `),V("suffix",`
 margin: 0 0 0 4px;
 font-size: var(--n-value-font-size);
 transition: .3s color var(--n-bezier);
 color: var(--n-value-suffix-text-color);
 `,[h("icon",{verticalAlign:"-0.125em"})])])]),Me=Object.assign(Object.assign({},M.props),{tabularNums:Boolean,label:String,value:[String,Number]}),W=T({name:"Statistic",props:Me,slots:Object,setup(e){const{mergedClsPrefixRef:s,inlineThemeDisabled:l,mergedRtlRef:r}=K(e),p=M("Statistic","-statistic",Oe,ve,e,s),o=he("Statistic",r,s),i=S(()=>{const{self:{labelFontWeight:u,valueFontSize:b,valueFontWeight:c,valuePrefixTextColor:m,labelTextColor:g,valueSuffixTextColor:C,valueTextColor:y,labelFontSize:f},common:{cubicBezierEaseInOut:z}}=p.value;return{"--n-bezier":z,"--n-label-font-size":f,"--n-label-font-weight":u,"--n-label-text-color":g,"--n-value-font-weight":c,"--n-value-font-size":b,"--n-value-prefix-text-color":m,"--n-value-suffix-text-color":C,"--n-value-text-color":y}}),a=l?U("statistic",void 0,i,e):void 0;return{rtlEnabled:o,mergedClsPrefix:s,cssVars:l?void 0:i,themeClass:a?.themeClass,onRender:a?.onRender}},render(){var e;const{mergedClsPrefix:s,$slots:{default:l,label:r,prefix:p,suffix:o}}=this;return(e=this.onRender)===null||e===void 0||e.call(this),t("div",{class:[`${s}-statistic`,this.themeClass,this.rtlEnabled&&`${s}-statistic--rtl`],style:this.cssVars},q(r,i=>t("div",{class:`${s}-statistic__label`},this.label||i)),t("div",{class:`${s}-statistic-value`,style:{fontVariantNumeric:this.tabularNums?"tabular-nums":""}},q(p,i=>i&&t("span",{class:`${s}-statistic-value__prefix`},i)),this.value!==void 0?t("span",{class:`${s}-statistic-value__content`},this.value):q(l,i=>i&&t("span",{class:`${s}-statistic-value__content`},i)),q(o,i=>i&&t("span",{class:`${s}-statistic-value__suffix`},i))))}}),Ae={summary(){return me("/dashboard/summary")}},Ve={class:"space-y-4"},qe={class:"space-y-3"},Ge={class:"flex justify-between text-sm"},Ee={class:"flex justify-between text-sm"},Le={class:"flex justify-between text-sm"},He={class:"text-sm text-slate-500"},Qe=T({__name:"DashboardView",setup(e){const{t:s}=ye(),l=be(),r=E(null),p=E(!1),o=E("");let i=null;async function a(){p.value=!0;try{r.value=await Ae.summary(),o.value=""}catch(c){o.value=c.message}finally{p.value=!1}}function u(c){if(!c)return"0 B";const m=["B","KB","MB","GB","TB"];let g=0,C=c;for(;C>=1024&&g<m.length-1;)C/=1024,g++;return`${C.toFixed(2)} ${m[g]}`}const b=S(()=>{const c=r.value?.tunnelCount;return c?c.tcp+c.udp+c.socks5+c.httpProxy+c.secret+c.p2p:0});return xe(()=>{l.isAdmin&&(a(),i=window.setInterval(a,5e3))}),Ce(()=>{i!==null&&window.clearInterval(i)}),(c,m)=>(G(),J("div",Ve,[n(l).isAdmin?(G(),J($e,{key:1},[d(n(We),{show:p.value&&!r.value},{default:v(()=>[d(n(ze),{cols:4,"x-gap":16,"y-gap":16,responsive:"screen","item-responsive":""},{default:v(()=>[d(n(I),{span:"1 m:1"},{default:v(()=>[d(n(R),null,{default:v(()=>[d(n(W),{label:n(s)("dashboard.clients")},{suffix:v(()=>[k(" / "+x(r.value?.clientCount??0),1)]),default:v(()=>[$("span",null,x(r.value?.clientOnlineCount??0),1)]),_:1},8,["label"])]),_:1})]),_:1}),d(n(I),{span:"1 m:1"},{default:v(()=>[d(n(R),null,{default:v(()=>[d(n(W),{label:n(s)("dashboard.hosts"),value:r.value?.hostCount??0},null,8,["label","value"])]),_:1})]),_:1}),d(n(I),{span:"1 m:1"},{default:v(()=>[d(n(R),null,{default:v(()=>[d(n(W),{label:n(s)("dashboard.tunnels"),value:b.value},null,8,["label","value"])]),_:1})]),_:1}),d(n(I),{span:"1 m:1"},{default:v(()=>[d(n(R),null,{default:v(()=>[d(n(W),{label:n(s)("dashboard.connections"),value:r.value?.connections??0},null,8,["label","value"])]),_:1})]),_:1}),d(n(I),{span:2},{default:v(()=>[d(n(R),{title:n(s)("dashboard.tunnelByMode")},{default:v(()=>[d(n(F),null,{default:v(()=>[d(n(D),{type:"info"},{default:v(()=>[k("TCP "+x(r.value?.tunnelCount.tcp??0),1)]),_:1}),d(n(D),{type:"info"},{default:v(()=>[k("UDP "+x(r.value?.tunnelCount.udp??0),1)]),_:1}),d(n(D),{type:"info"},{default:v(()=>[k("SOCKS5 "+x(r.value?.tunnelCount.socks5??0),1)]),_:1}),d(n(D),{type:"info"},{default:v(()=>[k("HTTP "+x(r.value?.tunnelCount.httpProxy??0),1)]),_:1}),d(n(D),{type:"success"},{default:v(()=>[k("SECRET "+x(r.value?.tunnelCount.secret??0),1)]),_:1}),d(n(D),{type:"success"},{default:v(()=>[k("P2P "+x(r.value?.tunnelCount.p2p??0),1)]),_:1})]),_:1})]),_:1},8,["title"])]),_:1}),d(n(I),{span:2},{default:v(()=>[d(n(R),{title:n(s)("dashboard.flow")},{default:v(()=>[d(n(F),null,{default:v(()=>[d(n(W),{label:n(s)("dashboard.inFlow"),value:u(r.value?.flow.in??0)},null,8,["label","value"]),d(n(W),{label:n(s)("dashboard.outFlow"),value:u(r.value?.flow.out??0)},null,8,["label","value"])]),_:1})]),_:1},8,["title"])]),_:1}),d(n(I),{span:2},{default:v(()=>[d(n(R),{title:n(s)("dashboard.system")},{default:v(()=>[$("div",qe,[$("div",null,[$("div",Ge,[m[0]||(m[0]=$("span",null,"CPU",-1)),$("span",null,x(Math.round(r.value?.system.cpu??0))+"%",1)]),d(n(X),{type:"line",percentage:Math.min(100,Math.round(r.value?.system.cpu??0)),"show-indicator":!1},null,8,["percentage"])]),$("div",null,[$("div",Ee,[m[1]||(m[1]=$("span",null,"MEM",-1)),$("span",null,x(Math.round(r.value?.system.mem??0))+"%",1)]),d(n(X),{type:"line",percentage:Math.min(100,Math.round(r.value?.system.mem??0)),"show-indicator":!1},null,8,["percentage"])]),$("div",null,[$("div",Le,[m[2]||(m[2]=$("span",null,"SWAP",-1)),$("span",null,x(Math.round(r.value?.system.swap??0))+"%",1)]),d(n(X),{type:"line",percentage:Math.min(100,Math.round(r.value?.system.swap??0)),"show-indicator":!1},null,8,["percentage"])]),$("div",He,x(n(s)("dashboard.load"))+": "+x(r.value?.load),1)])]),_:1},8,["title"])]),_:1}),d(n(I),{span:2},{default:v(()=>[d(n(R),{title:n(s)("dashboard.serverInfo")},{default:v(()=>[d(n(F),{vertical:""},{default:v(()=>[$("div",null,x(n(s)("dashboard.version"))+": "+x(r.value?.version),1),$("div",null,x(n(s)("dashboard.bridge"))+": "+x(r.value?.bridgeType)+" :"+x(r.value?.bridgePort),1),$("div",null,"HTTP Proxy: :"+x(r.value?.httpProxyPort||"-"),1),$("div",null,"HTTPS Proxy: :"+x(r.value?.httpsProxyPort||"-"),1),$("div",null,"P2P: "+x(r.value?.serverIp||"-")+":"+x(r.value?.p2pPort||"-"),1),$("div",null,"Log level: "+x(r.value?.logLevel||"-"),1)]),_:1})]),_:1},8,["title"])]),_:1})]),_:1})]),_:1},8,["show"]),o.value?(G(),Z(n(Q),{key:0,type:"error"},{default:v(()=>[k(x(o.value),1)]),_:1})):Se("",!0)],64)):(G(),Z(n(Q),{key:0,type:"info"},{default:v(()=>[k(x(n(s)("dashboard.userOnly")),1)]),_:1}))]))}});export{Qe as default};
