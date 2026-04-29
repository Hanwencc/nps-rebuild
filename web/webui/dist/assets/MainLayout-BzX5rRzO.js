import{d as B,m as u,n as Xe,s as Ze,p as Je,q as we,v as J,x as h,y as S,S as Ne,z as le,C as U,D as Te,E as ie,F as C,l as $,G as Y,H as v,I,J as He,K as D,L,M as ne,O as Z,P as Qe,Q as W,R as pe,T as ue,U as eo,V as se,W as oo,X as to,Y as Ie,Z as ro,_ as no,$ as lo,u as io,a as ao,a0 as so,a1 as co,w as F,f as N,o as uo,e as E,b as q,a2 as vo,a3 as ho,t as oe,k as te,B as mo,a4 as po,g as fo,h as go}from"./index-ByVz0pZH.js";import{C as bo,N as xo,a as Ae,V as Co,c as ce,b as yo,d as zo}from"./Dropdown-DUXcSf5k.js";import{f as de,u as ve}from"./Suffix-BZS8iX34.js";import{u as wo}from"./Tag-DrTi-tia.js";import{N as Io}from"./Space-EweXbsCT.js";import{N as So}from"./Switch-CVZaCLpc.js";import{_ as Ro}from"./_plugin-vue_export-helper-DlAUqK2U.js";import"./next-frame-once-C5Ksf8W7.js";const Po=B({name:"ChevronDownFilled",render(){return u("svg",{viewBox:"0 0 16 16",fill:"none",xmlns:"http://www.w3.org/2000/svg"},u("path",{d:"M3.20041 5.73966C3.48226 5.43613 3.95681 5.41856 4.26034 5.70041L8 9.22652L11.7397 5.70041C12.0432 5.41856 12.5177 5.43613 12.7996 5.73966C13.0815 6.0432 13.0639 6.51775 12.7603 6.7996L8.51034 10.7996C8.22258 11.0668 7.77743 11.0668 7.48967 10.7996L3.23966 6.7996C2.93613 6.51775 2.91856 6.0432 3.20041 5.73966Z",fill:"currentColor"}))}});function No(e){const{baseColor:o,textColor2:r,bodyColor:i,cardColor:a,dividerColor:l,actionColor:c,scrollbarColor:s,scrollbarColorHover:d,invertedColor:x}=e;return{textColor:r,textColorInverted:"#FFF",color:i,colorEmbedded:c,headerColor:a,headerColorInverted:x,footerColor:c,footerColorInverted:x,headerBorderColor:l,headerBorderColorInverted:x,footerBorderColor:l,footerBorderColorInverted:x,siderBorderColor:l,siderBorderColorInverted:x,siderColor:a,siderColorInverted:x,siderToggleButtonBorder:`1px solid ${l}`,siderToggleButtonColor:o,siderToggleButtonIconColor:r,siderToggleButtonIconColorInverted:r,siderToggleBarColor:we(i,s),siderToggleBarColorHover:we(i,d),__invertScrollbar:"true"}}const fe=Xe({name:"Layout",common:Je,peers:{Scrollbar:Ze},self:No}),ke=J("n-layout-sider"),ge={type:String,default:"static"},To=h("layout",`
 color: var(--n-text-color);
 background-color: var(--n-color);
 box-sizing: border-box;
 position: relative;
 z-index: auto;
 flex: auto;
 overflow: hidden;
 transition:
 box-shadow .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 color .3s var(--n-bezier);
`,[h("layout-scroll-container",`
 overflow-x: hidden;
 box-sizing: border-box;
 height: 100%;
 `),S("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),Ho={embedded:Boolean,position:ge,nativeScrollbar:{type:Boolean,default:!0},scrollbarProps:Object,onScroll:Function,contentClass:String,contentStyle:{type:[String,Object],default:""},hasSider:Boolean,siderPlacement:{type:String,default:"left"}},Me=J("n-layout");function _e(e){return B({name:e?"LayoutContent":"Layout",props:Object.assign(Object.assign({},U.props),Ho),setup(o){const r=$(null),i=$(null),{mergedClsPrefixRef:a,inlineThemeDisabled:l}=le(o),c=U("Layout","-layout",To,fe,o,a);function s(p,g){if(o.nativeScrollbar){const{value:R}=r;R&&(g===void 0?R.scrollTo(p):R.scrollTo(p,g))}else{const{value:R}=i;R&&R.scrollTo(p,g)}}Y(Me,o);let d=0,x=0;const M=p=>{var g;const R=p.target;d=R.scrollLeft,x=R.scrollTop,(g=o.onScroll)===null||g===void 0||g.call(o,p)};Te(()=>{if(o.nativeScrollbar){const p=r.value;p&&(p.scrollTop=x,p.scrollLeft=d)}});const k={display:"flex",flexWrap:"nowrap",width:"100%",flexDirection:"row"},f={scrollTo:s},H=C(()=>{const{common:{cubicBezierEaseInOut:p},self:g}=c.value;return{"--n-bezier":p,"--n-color":o.embedded?g.colorEmbedded:g.color,"--n-text-color":g.textColor}}),T=l?ie("layout",C(()=>o.embedded?"e":""),H,o):void 0;return Object.assign({mergedClsPrefix:a,scrollableElRef:r,scrollbarInstRef:i,hasSiderStyle:k,mergedTheme:c,handleNativeElScroll:M,cssVars:l?void 0:H,themeClass:T?.themeClass,onRender:T?.onRender},f)},render(){var o;const{mergedClsPrefix:r,hasSider:i}=this;(o=this.onRender)===null||o===void 0||o.call(this);const a=i?this.hasSiderStyle:void 0,l=[this.themeClass,e&&`${r}-layout-content`,`${r}-layout`,`${r}-layout--${this.position}-positioned`];return u("div",{class:l,style:this.cssVars},this.nativeScrollbar?u("div",{ref:"scrollableElRef",class:[`${r}-layout-scroll-container`,this.contentClass],style:[this.contentStyle,a],onScroll:this.handleNativeElScroll},this.$slots):u(Ne,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,contentClass:this.contentClass,contentStyle:[this.contentStyle,a]}),this.$slots))}})}const Se=_e(!1),Ao=_e(!0),ko=h("layout-header",`
 transition:
 color .3s var(--n-bezier),
 background-color .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 box-sizing: border-box;
 width: 100%;
 background-color: var(--n-color);
 color: var(--n-text-color);
`,[S("absolute-positioned",`
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 `),S("bordered",`
 border-bottom: solid 1px var(--n-border-color);
 `)]),Mo={position:ge,inverted:Boolean,bordered:{type:Boolean,default:!1}},_o=B({name:"LayoutHeader",props:Object.assign(Object.assign({},U.props),Mo),setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:r}=le(e),i=U("Layout","-layout-header",ko,fe,e,o),a=C(()=>{const{common:{cubicBezierEaseInOut:c},self:s}=i.value,d={"--n-bezier":c};return e.inverted?(d["--n-color"]=s.headerColorInverted,d["--n-text-color"]=s.textColorInverted,d["--n-border-color"]=s.headerBorderColorInverted):(d["--n-color"]=s.headerColor,d["--n-text-color"]=s.textColor,d["--n-border-color"]=s.headerBorderColor),d}),l=r?ie("layout-header",C(()=>e.inverted?"a":"b"),a,e):void 0;return{mergedClsPrefix:o,cssVars:r?void 0:a,themeClass:l?.themeClass,onRender:l?.onRender}},render(){var e;const{mergedClsPrefix:o}=this;return(e=this.onRender)===null||e===void 0||e.call(this),u("div",{class:[`${o}-layout-header`,this.themeClass,this.position&&`${o}-layout-header--${this.position}-positioned`,this.bordered&&`${o}-layout-header--bordered`],style:this.cssVars},this.$slots)}}),Bo=h("layout-sider",`
 flex-shrink: 0;
 box-sizing: border-box;
 position: relative;
 z-index: 1;
 color: var(--n-text-color);
 transition:
 color .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 min-width .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 transform .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 background-color: var(--n-color);
 display: flex;
 justify-content: flex-end;
`,[S("bordered",[v("border",`
 content: "";
 position: absolute;
 top: 0;
 bottom: 0;
 width: 1px;
 background-color: var(--n-border-color);
 transition: background-color .3s var(--n-bezier);
 `)]),v("left-placement",[S("bordered",[v("border",`
 right: 0;
 `)])]),S("right-placement",`
 justify-content: flex-start;
 `,[S("bordered",[v("border",`
 left: 0;
 `)]),S("collapsed",[h("layout-toggle-button",[h("base-icon",`
 transform: rotate(180deg);
 `)]),h("layout-toggle-bar",[I("&:hover",[v("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])])]),h("layout-toggle-button",`
 left: 0;
 transform: translateX(-50%) translateY(-50%);
 `,[h("base-icon",`
 transform: rotate(0);
 `)]),h("layout-toggle-bar",`
 left: -28px;
 transform: rotate(180deg);
 `,[I("&:hover",[v("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})])])]),S("collapsed",[h("layout-toggle-bar",[I("&:hover",[v("top",{transform:"rotate(-12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(12deg) scale(1.15) translateY(2px)"})])]),h("layout-toggle-button",[h("base-icon",`
 transform: rotate(0);
 `)])]),h("layout-toggle-button",`
 transition:
 color .3s var(--n-bezier),
 right .3s var(--n-bezier),
 left .3s var(--n-bezier),
 border-color .3s var(--n-bezier),
 background-color .3s var(--n-bezier);
 cursor: pointer;
 width: 24px;
 height: 24px;
 position: absolute;
 top: 50%;
 right: 0;
 border-radius: 50%;
 display: flex;
 align-items: center;
 justify-content: center;
 font-size: 18px;
 color: var(--n-toggle-button-icon-color);
 border: var(--n-toggle-button-border);
 background-color: var(--n-toggle-button-color);
 box-shadow: 0 2px 4px 0px rgba(0, 0, 0, .06);
 transform: translateX(50%) translateY(-50%);
 z-index: 1;
 `,[h("base-icon",`
 transition: transform .3s var(--n-bezier);
 transform: rotate(180deg);
 `)]),h("layout-toggle-bar",`
 cursor: pointer;
 height: 72px;
 width: 32px;
 position: absolute;
 top: calc(50% - 36px);
 right: -28px;
 `,[v("top, bottom",`
 position: absolute;
 width: 4px;
 border-radius: 2px;
 height: 38px;
 left: 14px;
 transition: 
 background-color .3s var(--n-bezier),
 transform .3s var(--n-bezier);
 `),v("bottom",`
 position: absolute;
 top: 34px;
 `),I("&:hover",[v("top",{transform:"rotate(12deg) scale(1.15) translateY(-2px)"}),v("bottom",{transform:"rotate(-12deg) scale(1.15) translateY(2px)"})]),v("top, bottom",{backgroundColor:"var(--n-toggle-bar-color)"}),I("&:hover",[v("top, bottom",{backgroundColor:"var(--n-toggle-bar-color-hover)"})])]),v("border",`
 position: absolute;
 top: 0;
 right: 0;
 bottom: 0;
 width: 1px;
 transition: background-color .3s var(--n-bezier);
 `),h("layout-sider-scroll-container",`
 flex-grow: 1;
 flex-shrink: 0;
 box-sizing: border-box;
 height: 100%;
 opacity: 0;
 transition: opacity .3s var(--n-bezier);
 max-width: 100%;
 `),S("show-content",[h("layout-sider-scroll-container",{opacity:1})]),S("absolute-positioned",`
 position: absolute;
 left: 0;
 top: 0;
 bottom: 0;
 `)]),Oo=B({props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return u("div",{onClick:this.onClick,class:`${e}-layout-toggle-bar`},u("div",{class:`${e}-layout-toggle-bar__top`}),u("div",{class:`${e}-layout-toggle-bar__bottom`}))}}),Eo=B({name:"LayoutToggleButton",props:{clsPrefix:{type:String,required:!0},onClick:Function},render(){const{clsPrefix:e}=this;return u("div",{class:`${e}-layout-toggle-button`,onClick:this.onClick},u(He,{clsPrefix:e},{default:()=>u(bo,null)}))}}),Lo={position:ge,bordered:Boolean,collapsedWidth:{type:Number,default:48},width:{type:[Number,String],default:272},contentClass:String,contentStyle:{type:[String,Object],default:""},collapseMode:{type:String,default:"transform"},collapsed:{type:Boolean,default:void 0},defaultCollapsed:Boolean,showCollapsedContent:{type:Boolean,default:!0},showTrigger:{type:[Boolean,String],default:!1},nativeScrollbar:{type:Boolean,default:!0},inverted:Boolean,scrollbarProps:Object,triggerClass:String,triggerStyle:[String,Object],collapsedTriggerClass:String,collapsedTriggerStyle:[String,Object],"onUpdate:collapsed":[Function,Array],onUpdateCollapsed:[Function,Array],onAfterEnter:Function,onAfterLeave:Function,onExpand:[Function,Array],onCollapse:[Function,Array],onScroll:Function},$o=B({name:"LayoutSider",props:Object.assign(Object.assign({},U.props),Lo),setup(e){const o=D(Me),r=$(null),i=$(null),a=$(e.defaultCollapsed),l=ve(ne(e,"collapsed"),a),c=C(()=>de(l.value?e.collapsedWidth:e.width)),s=C(()=>e.collapseMode!=="transform"?{}:{minWidth:de(e.width)}),d=C(()=>o?o.siderPlacement:"left");function x(A,z){if(e.nativeScrollbar){const{value:w}=r;w&&(z===void 0?w.scrollTo(A):w.scrollTo(A,z))}else{const{value:w}=i;w&&w.scrollTo(A,z)}}function M(){const{"onUpdate:collapsed":A,onUpdateCollapsed:z,onExpand:w,onCollapse:K}=e,{value:V}=l;z&&L(z,!V),A&&L(A,!V),a.value=!V,V?w&&L(w):K&&L(K)}let k=0,f=0;const H=A=>{var z;const w=A.target;k=w.scrollLeft,f=w.scrollTop,(z=e.onScroll)===null||z===void 0||z.call(e,A)};Te(()=>{if(e.nativeScrollbar){const A=r.value;A&&(A.scrollTop=f,A.scrollLeft=k)}}),Y(ke,{collapsedRef:l,collapseModeRef:ne(e,"collapseMode")});const{mergedClsPrefixRef:T,inlineThemeDisabled:p}=le(e),g=U("Layout","-layout-sider",Bo,fe,e,T);function R(A){var z,w;A.propertyName==="max-width"&&(l.value?(z=e.onAfterLeave)===null||z===void 0||z.call(e):(w=e.onAfterEnter)===null||w===void 0||w.call(e))}const X={scrollTo:x},j=C(()=>{const{common:{cubicBezierEaseInOut:A},self:z}=g.value,{siderToggleButtonColor:w,siderToggleButtonBorder:K,siderToggleBarColor:V,siderToggleBarColorHover:ae}=z,_={"--n-bezier":A,"--n-toggle-button-color":w,"--n-toggle-button-border":K,"--n-toggle-bar-color":V,"--n-toggle-bar-color-hover":ae};return e.inverted?(_["--n-color"]=z.siderColorInverted,_["--n-text-color"]=z.textColorInverted,_["--n-border-color"]=z.siderBorderColorInverted,_["--n-toggle-button-icon-color"]=z.siderToggleButtonIconColorInverted,_.__invertScrollbar=z.__invertScrollbar):(_["--n-color"]=z.siderColor,_["--n-text-color"]=z.textColor,_["--n-border-color"]=z.siderBorderColor,_["--n-toggle-button-icon-color"]=z.siderToggleButtonIconColor),_}),O=p?ie("layout-sider",C(()=>e.inverted?"a":"b"),j,e):void 0;return Object.assign({scrollableElRef:r,scrollbarInstRef:i,mergedClsPrefix:T,mergedTheme:g,styleMaxWidth:c,mergedCollapsed:l,scrollContainerStyle:s,siderPlacement:d,handleNativeElScroll:H,handleTransitionend:R,handleTriggerClick:M,inlineThemeDisabled:p,cssVars:j,themeClass:O?.themeClass,onRender:O?.onRender},X)},render(){var e;const{mergedClsPrefix:o,mergedCollapsed:r,showTrigger:i}=this;return(e=this.onRender)===null||e===void 0||e.call(this),u("aside",{class:[`${o}-layout-sider`,this.themeClass,`${o}-layout-sider--${this.position}-positioned`,`${o}-layout-sider--${this.siderPlacement}-placement`,this.bordered&&`${o}-layout-sider--bordered`,r&&`${o}-layout-sider--collapsed`,(!r||this.showCollapsedContent)&&`${o}-layout-sider--show-content`],onTransitionend:this.handleTransitionend,style:[this.inlineThemeDisabled?void 0:this.cssVars,{maxWidth:this.styleMaxWidth,width:de(this.width)}]},this.nativeScrollbar?u("div",{class:[`${o}-layout-sider-scroll-container`,this.contentClass],onScroll:this.handleNativeElScroll,style:[this.scrollContainerStyle,{overflow:"auto"},this.contentStyle],ref:"scrollableElRef"},this.$slots):u(Ne,Object.assign({},this.scrollbarProps,{onScroll:this.onScroll,ref:"scrollbarInstRef",style:this.scrollContainerStyle,contentStyle:this.contentStyle,contentClass:this.contentClass,theme:this.mergedTheme.peers.Scrollbar,themeOverrides:this.mergedTheme.peerOverrides.Scrollbar,builtinThemeOverrides:this.inverted&&this.cssVars.__invertScrollbar==="true"?{colorHover:"rgba(255, 255, 255, .4)",color:"rgba(255, 255, 255, .3)"}:void 0}),this.$slots),i?i==="bar"?u(Oo,{clsPrefix:o,class:r?this.collapsedTriggerClass:this.triggerClass,style:r?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):u(Eo,{clsPrefix:o,class:r?this.collapsedTriggerClass:this.triggerClass,style:r?this.collapsedTriggerStyle:this.triggerStyle,onClick:this.handleTriggerClick}):null,this.bordered?u("div",{class:`${o}-layout-sider__border`}):null)}}),Q=J("n-menu"),Be=J("n-submenu"),be=J("n-menu-item-group"),Re=[I("&::before","background-color: var(--n-item-color-hover);"),v("arrow",`
 color: var(--n-arrow-color-hover);
 `),v("icon",`
 color: var(--n-item-icon-color-hover);
 `),h("menu-item-content-header",`
 color: var(--n-item-text-color-hover);
 `,[I("a",`
 color: var(--n-item-text-color-hover);
 `),v("extra",`
 color: var(--n-item-text-color-hover);
 `)])],Pe=[v("icon",`
 color: var(--n-item-icon-color-hover-horizontal);
 `),h("menu-item-content-header",`
 color: var(--n-item-text-color-hover-horizontal);
 `,[I("a",`
 color: var(--n-item-text-color-hover-horizontal);
 `),v("extra",`
 color: var(--n-item-text-color-hover-horizontal);
 `)])],Fo=I([h("menu",`
 background-color: var(--n-color);
 color: var(--n-item-text-color);
 overflow: hidden;
 transition: background-color .3s var(--n-bezier);
 box-sizing: border-box;
 font-size: var(--n-font-size);
 padding-bottom: 6px;
 `,[S("horizontal",`
 max-width: 100%;
 width: 100%;
 display: flex;
 overflow: hidden;
 padding-bottom: 0;
 `,[h("submenu","margin: 0;"),h("menu-item","margin: 0;"),h("menu-item-content",`
 padding: 0 20px;
 border-bottom: 2px solid #0000;
 `,[I("&::before","display: none;"),S("selected","border-bottom: 2px solid var(--n-border-color-horizontal)")]),h("menu-item-content",[S("selected",[v("icon","color: var(--n-item-icon-color-active-horizontal);"),h("menu-item-content-header",`
 color: var(--n-item-text-color-active-horizontal);
 `,[I("a","color: var(--n-item-text-color-active-horizontal);"),v("extra","color: var(--n-item-text-color-active-horizontal);")])]),S("child-active",`
 border-bottom: 2px solid var(--n-border-color-horizontal);
 `,[h("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-horizontal);
 `,[I("a",`
 color: var(--n-item-text-color-child-active-horizontal);
 `),v("extra",`
 color: var(--n-item-text-color-child-active-horizontal);
 `)]),v("icon",`
 color: var(--n-item-icon-color-child-active-horizontal);
 `)]),Z("disabled",[Z("selected, child-active",[I("&:focus-within",Pe)]),S("selected",[G(null,[v("icon","color: var(--n-item-icon-color-active-hover-horizontal);"),h("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover-horizontal);
 `,[I("a","color: var(--n-item-text-color-active-hover-horizontal);"),v("extra","color: var(--n-item-text-color-active-hover-horizontal);")])])]),S("child-active",[G(null,[v("icon","color: var(--n-item-icon-color-child-active-hover-horizontal);"),h("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover-horizontal);
 `,[I("a","color: var(--n-item-text-color-child-active-hover-horizontal);"),v("extra","color: var(--n-item-text-color-child-active-hover-horizontal);")])])]),G("border-bottom: 2px solid var(--n-border-color-horizontal);",Pe)]),h("menu-item-content-header",[I("a","color: var(--n-item-text-color-horizontal);")])])]),Z("responsive",[h("menu-item-content-header",`
 overflow: hidden;
 text-overflow: ellipsis;
 `)]),S("collapsed",[h("menu-item-content",[S("selected",[I("&::before",`
 background-color: var(--n-item-color-active-collapsed) !important;
 `)]),h("menu-item-content-header","opacity: 0;"),v("arrow","opacity: 0;"),v("icon","color: var(--n-item-icon-color-collapsed);")])]),h("menu-item",`
 height: var(--n-item-height);
 margin-top: 6px;
 position: relative;
 `),h("menu-item-content",`
 box-sizing: border-box;
 line-height: 1.75;
 height: 100%;
 display: grid;
 grid-template-areas: "icon content arrow";
 grid-template-columns: auto 1fr auto;
 align-items: center;
 cursor: pointer;
 position: relative;
 padding-right: 18px;
 transition:
 background-color .3s var(--n-bezier),
 padding-left .3s var(--n-bezier),
 border-color .3s var(--n-bezier);
 `,[I("> *","z-index: 1;"),I("&::before",`
 z-index: auto;
 content: "";
 background-color: #0000;
 position: absolute;
 left: 8px;
 right: 8px;
 top: 0;
 bottom: 0;
 pointer-events: none;
 border-radius: var(--n-border-radius);
 transition: background-color .3s var(--n-bezier);
 `),S("disabled",`
 opacity: .45;
 cursor: not-allowed;
 `),S("collapsed",[v("arrow","transform: rotate(0);")]),S("selected",[I("&::before","background-color: var(--n-item-color-active);"),v("arrow","color: var(--n-arrow-color-active);"),v("icon","color: var(--n-item-icon-color-active);"),h("menu-item-content-header",`
 color: var(--n-item-text-color-active);
 `,[I("a","color: var(--n-item-text-color-active);"),v("extra","color: var(--n-item-text-color-active);")])]),S("child-active",[h("menu-item-content-header",`
 color: var(--n-item-text-color-child-active);
 `,[I("a",`
 color: var(--n-item-text-color-child-active);
 `),v("extra",`
 color: var(--n-item-text-color-child-active);
 `)]),v("arrow",`
 color: var(--n-arrow-color-child-active);
 `),v("icon",`
 color: var(--n-item-icon-color-child-active);
 `)]),Z("disabled",[Z("selected, child-active",[I("&:focus-within",Re)]),S("selected",[G(null,[v("arrow","color: var(--n-arrow-color-active-hover);"),v("icon","color: var(--n-item-icon-color-active-hover);"),h("menu-item-content-header",`
 color: var(--n-item-text-color-active-hover);
 `,[I("a","color: var(--n-item-text-color-active-hover);"),v("extra","color: var(--n-item-text-color-active-hover);")])])]),S("child-active",[G(null,[v("arrow","color: var(--n-arrow-color-child-active-hover);"),v("icon","color: var(--n-item-icon-color-child-active-hover);"),h("menu-item-content-header",`
 color: var(--n-item-text-color-child-active-hover);
 `,[I("a","color: var(--n-item-text-color-child-active-hover);"),v("extra","color: var(--n-item-text-color-child-active-hover);")])])]),S("selected",[G(null,[I("&::before","background-color: var(--n-item-color-active-hover);")])]),G(null,Re)]),v("icon",`
 grid-area: icon;
 color: var(--n-item-icon-color);
 transition:
 color .3s var(--n-bezier),
 font-size .3s var(--n-bezier),
 margin-right .3s var(--n-bezier);
 box-sizing: content-box;
 display: inline-flex;
 align-items: center;
 justify-content: center;
 `),v("arrow",`
 grid-area: arrow;
 font-size: 16px;
 color: var(--n-arrow-color);
 transform: rotate(180deg);
 opacity: 1;
 transition:
 color .3s var(--n-bezier),
 transform 0.2s var(--n-bezier),
 opacity 0.2s var(--n-bezier);
 `),h("menu-item-content-header",`
 grid-area: content;
 transition:
 color .3s var(--n-bezier),
 opacity .3s var(--n-bezier);
 opacity: 1;
 white-space: nowrap;
 color: var(--n-item-text-color);
 `,[I("a",`
 outline: none;
 text-decoration: none;
 transition: color .3s var(--n-bezier);
 color: var(--n-item-text-color);
 `,[I("&::before",`
 content: "";
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 `)]),v("extra",`
 font-size: .93em;
 color: var(--n-group-text-color);
 transition: color .3s var(--n-bezier);
 `)])]),h("submenu",`
 cursor: pointer;
 position: relative;
 margin-top: 6px;
 `,[h("menu-item-content",`
 height: var(--n-item-height);
 `),h("submenu-children",`
 overflow: hidden;
 padding: 0;
 `,[Qe({duration:".2s"})])]),h("menu-item-group",[h("menu-item-group-title",`
 margin-top: 6px;
 color: var(--n-group-text-color);
 cursor: default;
 font-size: .93em;
 height: 36px;
 display: flex;
 align-items: center;
 transition:
 padding-left .3s var(--n-bezier),
 color .3s var(--n-bezier);
 `)])]),h("menu-tooltip",[I("a",`
 color: inherit;
 text-decoration: none;
 `)]),h("menu-divider",`
 transition: background-color .3s var(--n-bezier);
 background-color: var(--n-divider-color);
 height: 1px;
 margin: 6px 18px;
 `)]);function G(e,o){return[S("hover",e,o),I("&:hover",e,o)]}const Oe=B({name:"MenuOptionContent",props:{collapsed:Boolean,disabled:Boolean,title:[String,Function],icon:Function,extra:[String,Function],showArrow:Boolean,childActive:Boolean,hover:Boolean,paddingLeft:Number,selected:Boolean,maxIconSize:{type:Number,required:!0},activeIconSize:{type:Number,required:!0},iconMarginRight:{type:Number,required:!0},clsPrefix:{type:String,required:!0},onClick:Function,tmNode:{type:Object,required:!0},isEllipsisPlaceholder:Boolean},setup(e){const{props:o}=D(Q);return{menuProps:o,style:C(()=>{const{paddingLeft:r}=e;return{paddingLeft:r&&`${r}px`}}),iconStyle:C(()=>{const{maxIconSize:r,activeIconSize:i,iconMarginRight:a}=e;return{width:`${r}px`,height:`${r}px`,fontSize:`${i}px`,marginRight:`${a}px`}})}},render(){const{clsPrefix:e,tmNode:o,menuProps:{renderIcon:r,renderLabel:i,renderExtra:a,expandIcon:l}}=this,c=r?r(o.rawNode):W(this.icon);return u("div",{onClick:s=>{var d;(d=this.onClick)===null||d===void 0||d.call(this,s)},role:"none",class:[`${e}-menu-item-content`,{[`${e}-menu-item-content--selected`]:this.selected,[`${e}-menu-item-content--collapsed`]:this.collapsed,[`${e}-menu-item-content--child-active`]:this.childActive,[`${e}-menu-item-content--disabled`]:this.disabled,[`${e}-menu-item-content--hover`]:this.hover}],style:this.style},c&&u("div",{class:`${e}-menu-item-content__icon`,style:this.iconStyle,role:"none"},[c]),u("div",{class:`${e}-menu-item-content-header`,role:"none"},this.isEllipsisPlaceholder?this.title:i?i(o.rawNode):W(this.title),this.extra||a?u("span",{class:`${e}-menu-item-content-header__extra`}," ",a?a(o.rawNode):W(this.extra)):null),this.showArrow?u(He,{ariaHidden:!0,class:`${e}-menu-item-content__arrow`,clsPrefix:e},{default:()=>l?l(o.rawNode):u(Po,null)}):null)}}),re=8;function xe(e){const o=D(Q),{props:r,mergedCollapsedRef:i}=o,a=D(Be,null),l=D(be,null),c=C(()=>r.mode==="horizontal"),s=C(()=>c.value?r.dropdownPlacement:"tmNodes"in e?"right-start":"right"),d=C(()=>{var f;return Math.max((f=r.collapsedIconSize)!==null&&f!==void 0?f:r.iconSize,r.iconSize)}),x=C(()=>{var f;return!c.value&&e.root&&i.value&&(f=r.collapsedIconSize)!==null&&f!==void 0?f:r.iconSize}),M=C(()=>{if(c.value)return;const{collapsedWidth:f,indent:H,rootIndent:T}=r,{root:p,isGroup:g}=e,R=T===void 0?H:T;return p?i.value?f/2-d.value/2:R:l&&typeof l.paddingLeftRef.value=="number"?H/2+l.paddingLeftRef.value:a&&typeof a.paddingLeftRef.value=="number"?(g?H/2:H)+a.paddingLeftRef.value:0}),k=C(()=>{const{collapsedWidth:f,indent:H,rootIndent:T}=r,{value:p}=d,{root:g}=e;return c.value||!g||!i.value?re:(T===void 0?H:T)+p+re-(f+p)/2});return{dropdownPlacement:s,activeIconSize:x,maxIconSize:d,paddingLeft:M,iconMarginRight:k,NMenu:o,NSubmenu:a,NMenuOptionGroup:l}}const Ce={internalKey:{type:[String,Number],required:!0},root:Boolean,isGroup:Boolean,level:{type:Number,required:!0},title:[String,Function],extra:[String,Function]},Vo=B({name:"MenuDivider",setup(){const e=D(Q),{mergedClsPrefixRef:o,isHorizontalRef:r}=e;return()=>r.value?null:u("div",{class:`${o.value}-menu-divider`})}}),Ee=Object.assign(Object.assign({},Ce),{tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function}),jo=pe(Ee),Ko=B({name:"MenuOption",props:Ee,setup(e){const o=xe(e),{NSubmenu:r,NMenu:i,NMenuOptionGroup:a}=o,{props:l,mergedClsPrefixRef:c,mergedCollapsedRef:s}=i,d=r?r.mergedDisabledRef:a?a.mergedDisabledRef:{value:!1},x=C(()=>d.value||e.disabled);function M(f){const{onClick:H}=e;H&&H(f)}function k(f){x.value||(i.doSelect(e.internalKey,e.tmNode.rawNode),M(f))}return{mergedClsPrefix:c,dropdownPlacement:o.dropdownPlacement,paddingLeft:o.paddingLeft,iconMarginRight:o.iconMarginRight,maxIconSize:o.maxIconSize,activeIconSize:o.activeIconSize,mergedTheme:i.mergedThemeRef,menuProps:l,dropdownEnabled:ue(()=>e.root&&s.value&&l.mode!=="horizontal"&&!x.value),selected:ue(()=>i.mergedValueRef.value===e.internalKey),mergedDisabled:x,handleClick:k}},render(){const{mergedClsPrefix:e,mergedTheme:o,tmNode:r,menuProps:{renderLabel:i,nodeProps:a}}=this,l=a?.(r.rawNode);return u("div",Object.assign({},l,{role:"menuitem",class:[`${e}-menu-item`,l?.class]}),u(xo,{theme:o.peers.Tooltip,themeOverrides:o.peerOverrides.Tooltip,trigger:"hover",placement:this.dropdownPlacement,disabled:!this.dropdownEnabled||this.title===void 0,internalExtraClass:["menu-tooltip"]},{default:()=>i?i(r.rawNode):W(this.title),trigger:()=>u(Oe,{tmNode:r,clsPrefix:e,paddingLeft:this.paddingLeft,iconMarginRight:this.iconMarginRight,maxIconSize:this.maxIconSize,activeIconSize:this.activeIconSize,selected:this.selected,title:this.title,extra:this.extra,disabled:this.mergedDisabled,icon:this.icon,onClick:this.handleClick})}))}}),Le=Object.assign(Object.assign({},Ce),{tmNode:{type:Object,required:!0},tmNodes:{type:Array,required:!0}}),Do=pe(Le),Uo=B({name:"MenuOptionGroup",props:Le,setup(e){const o=xe(e),{NSubmenu:r}=o,i=C(()=>r?.mergedDisabledRef.value?!0:e.tmNode.disabled);Y(be,{paddingLeftRef:o.paddingLeft,mergedDisabledRef:i});const{mergedClsPrefixRef:a,props:l}=D(Q);return function(){const{value:c}=a,s=o.paddingLeft.value,{nodeProps:d}=l,x=d?.(e.tmNode.rawNode);return u("div",{class:`${c}-menu-item-group`,role:"group"},u("div",Object.assign({},x,{class:[`${c}-menu-item-group-title`,x?.class],style:[x?.style||"",s!==void 0?`padding-left: ${s}px;`:""]}),W(e.title),e.extra?u(eo,null," ",W(e.extra)):null),u("div",null,e.tmNodes.map(M=>ye(M,l))))}}});function he(e){return e.type==="divider"||e.type==="render"}function Go(e){return e.type==="divider"}function ye(e,o){const{rawNode:r}=e,{show:i}=r;if(i===!1)return null;if(he(r))return Go(r)?u(Vo,Object.assign({key:e.key},r.props)):null;const{labelField:a}=o,{key:l,level:c,isGroup:s}=e,d=Object.assign(Object.assign({},r),{title:r.title||r[a],extra:r.titleExtra||r.extra,key:l,internalKey:l,level:c,root:c===0,isGroup:s});return e.children?e.isGroup?u(Uo,se(d,Do,{tmNode:e,tmNodes:e.children,key:l})):u(me,se(d,qo,{key:l,rawNodes:r[o.childrenField],tmNodes:e.children,tmNode:e})):u(Ko,se(d,jo,{key:l,tmNode:e}))}const $e=Object.assign(Object.assign({},Ce),{rawNodes:{type:Array,default:()=>[]},tmNodes:{type:Array,default:()=>[]},tmNode:{type:Object,required:!0},disabled:Boolean,icon:Function,onClick:Function,domId:String,virtualChildActive:{type:Boolean,default:void 0},isEllipsisPlaceholder:Boolean}),qo=pe($e),me=B({name:"Submenu",props:$e,setup(e){const o=xe(e),{NMenu:r,NSubmenu:i}=o,{props:a,mergedCollapsedRef:l,mergedThemeRef:c}=r,s=C(()=>{const{disabled:f}=e;return i?.mergedDisabledRef.value||a.disabled?!0:f}),d=$(!1);Y(Be,{paddingLeftRef:o.paddingLeft,mergedDisabledRef:s}),Y(be,null);function x(){const{onClick:f}=e;f&&f()}function M(){s.value||(l.value||r.toggleExpand(e.internalKey),x())}function k(f){d.value=f}return{menuProps:a,mergedTheme:c,doSelect:r.doSelect,inverted:r.invertedRef,isHorizontal:r.isHorizontalRef,mergedClsPrefix:r.mergedClsPrefixRef,maxIconSize:o.maxIconSize,activeIconSize:o.activeIconSize,iconMarginRight:o.iconMarginRight,dropdownPlacement:o.dropdownPlacement,dropdownShow:d,paddingLeft:o.paddingLeft,mergedDisabled:s,mergedValue:r.mergedValueRef,childActive:ue(()=>{var f;return(f=e.virtualChildActive)!==null&&f!==void 0?f:r.activePathRef.value.includes(e.internalKey)}),collapsed:C(()=>a.mode==="horizontal"?!1:l.value?!0:!r.mergedExpandedKeysRef.value.includes(e.internalKey)),dropdownEnabled:C(()=>!s.value&&(a.mode==="horizontal"||l.value)),handlePopoverShowChange:k,handleClick:M}},render(){var e;const{mergedClsPrefix:o,menuProps:{renderIcon:r,renderLabel:i}}=this,a=()=>{const{isHorizontal:c,paddingLeft:s,collapsed:d,mergedDisabled:x,maxIconSize:M,activeIconSize:k,title:f,childActive:H,icon:T,handleClick:p,menuProps:{nodeProps:g},dropdownShow:R,iconMarginRight:X,tmNode:j,mergedClsPrefix:O,isEllipsisPlaceholder:A,extra:z}=this,w=g?.(j.rawNode);return u("div",Object.assign({},w,{class:[`${O}-menu-item`,w?.class],role:"menuitem"}),u(Oe,{tmNode:j,paddingLeft:s,collapsed:d,disabled:x,iconMarginRight:X,maxIconSize:M,activeIconSize:k,title:f,extra:z,showArrow:!c,childActive:H,clsPrefix:O,icon:T,hover:R,onClick:p,isEllipsisPlaceholder:A}))},l=()=>u(oo,null,{default:()=>{const{tmNodes:c,collapsed:s}=this;return s?null:u("div",{class:`${o}-submenu-children`,role:"menu"},c.map(d=>ye(d,this.menuProps)))}});return this.root?u(Ae,Object.assign({size:"large",trigger:"hover"},(e=this.menuProps)===null||e===void 0?void 0:e.dropdownProps,{themeOverrides:this.mergedTheme.peerOverrides.Dropdown,theme:this.mergedTheme.peers.Dropdown,builtinThemeOverrides:{fontSizeLarge:"14px",optionIconSizeLarge:"18px"},value:this.mergedValue,disabled:!this.dropdownEnabled,placement:this.dropdownPlacement,keyField:this.menuProps.keyField,labelField:this.menuProps.labelField,childrenField:this.menuProps.childrenField,onUpdateShow:this.handlePopoverShowChange,options:this.rawNodes,onSelect:this.doSelect,inverted:this.inverted,renderIcon:r,renderLabel:i}),{default:()=>u("div",{class:`${o}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},a(),this.isHorizontal?null:l())}):u("div",{class:`${o}-submenu`,role:"menu","aria-expanded":!this.collapsed,id:this.domId},a(),l())}}),Wo=Object.assign(Object.assign({},U.props),{options:{type:Array,default:()=>[]},collapsed:{type:Boolean,default:void 0},collapsedWidth:{type:Number,default:48},iconSize:{type:Number,default:20},collapsedIconSize:{type:Number,default:24},rootIndent:Number,indent:{type:Number,default:32},labelField:{type:String,default:"label"},keyField:{type:String,default:"key"},childrenField:{type:String,default:"children"},disabledField:{type:String,default:"disabled"},defaultExpandAll:Boolean,defaultExpandedKeys:Array,expandedKeys:Array,value:[String,Number],defaultValue:{type:[String,Number],default:null},mode:{type:String,default:"vertical"},watchProps:{type:Array,default:void 0},disabled:Boolean,show:{type:Boolean,default:!0},inverted:Boolean,"onUpdate:expandedKeys":[Function,Array],onUpdateExpandedKeys:[Function,Array],onUpdateValue:[Function,Array],"onUpdate:value":[Function,Array],expandIcon:Function,renderIcon:Function,renderLabel:Function,renderExtra:Function,dropdownProps:Object,accordion:Boolean,nodeProps:Function,dropdownPlacement:{type:String,default:"bottom"},responsive:Boolean,items:Array,onOpenNamesChange:[Function,Array],onSelect:[Function,Array],onExpandedNamesChange:[Function,Array],expandedNames:Array,defaultExpandedNames:Array}),Yo=B({name:"Menu",inheritAttrs:!1,props:Wo,setup(e){const{mergedClsPrefixRef:o,inlineThemeDisabled:r}=le(e),i=U("Menu","-menu",Fo,lo,e,o),a=D(ke,null),l=C(()=>{var m;const{collapsed:y}=e;if(y!==void 0)return y;if(a){const{collapseModeRef:t,collapsedRef:b}=a;if(t.value==="width")return(m=b.value)!==null&&m!==void 0?m:!1}return!1}),c=C(()=>{const{keyField:m,childrenField:y,disabledField:t}=e;return ce(e.items||e.options,{getIgnored(b){return he(b)},getChildren(b){return b[y]},getDisabled(b){return b[t]},getKey(b){var P;return(P=b[m])!==null&&P!==void 0?P:b.name}})}),s=C(()=>new Set(c.value.treeNodes.map(m=>m.key))),{watchProps:d}=e,x=$(null);d?.includes("defaultValue")?Ie(()=>{x.value=e.defaultValue}):x.value=e.defaultValue;const M=ne(e,"value"),k=ve(M,x),f=$([]),H=()=>{f.value=e.defaultExpandAll?c.value.getNonLeafKeys():e.defaultExpandedNames||e.defaultExpandedKeys||c.value.getPath(k.value,{includeSelf:!1}).keyPath};d?.includes("defaultExpandedKeys")?Ie(H):H();const T=wo(e,["expandedNames","expandedKeys"]),p=ve(T,f),g=C(()=>c.value.treeNodes),R=C(()=>c.value.getPath(k.value).keyPath);Y(Q,{props:e,mergedCollapsedRef:l,mergedThemeRef:i,mergedValueRef:k,mergedExpandedKeysRef:p,activePathRef:R,mergedClsPrefixRef:o,isHorizontalRef:C(()=>e.mode==="horizontal"),invertedRef:ne(e,"inverted"),doSelect:X,toggleExpand:O});function X(m,y){const{"onUpdate:value":t,onUpdateValue:b,onSelect:P}=e;b&&L(b,m,y),t&&L(t,m,y),P&&L(P,m,y),x.value=m}function j(m){const{"onUpdate:expandedKeys":y,onUpdateExpandedKeys:t,onExpandedNamesChange:b,onOpenNamesChange:P}=e;y&&L(y,m),t&&L(t,m),b&&L(b,m),P&&L(P,m),f.value=m}function O(m){const y=Array.from(p.value),t=y.findIndex(b=>b===m);if(~t)y.splice(t,1);else{if(e.accordion&&s.value.has(m)){const b=y.findIndex(P=>s.value.has(P));b>-1&&y.splice(b,1)}y.push(m)}j(y)}const A=m=>{const y=c.value.getPath(m??k.value,{includeSelf:!1}).keyPath;if(!y.length)return;const t=Array.from(p.value),b=new Set([...t,...y]);e.accordion&&s.value.forEach(P=>{b.has(P)&&!y.includes(P)&&b.delete(P)}),j(Array.from(b))},z=C(()=>{const{inverted:m}=e,{common:{cubicBezierEaseInOut:y},self:t}=i.value,{borderRadius:b,borderColorHorizontal:P,fontSize:qe,itemHeight:We,dividerColor:Ye}=t,n={"--n-divider-color":Ye,"--n-bezier":y,"--n-font-size":qe,"--n-border-color-horizontal":P,"--n-border-radius":b,"--n-item-height":We};return m?(n["--n-group-text-color"]=t.groupTextColorInverted,n["--n-color"]=t.colorInverted,n["--n-item-text-color"]=t.itemTextColorInverted,n["--n-item-text-color-hover"]=t.itemTextColorHoverInverted,n["--n-item-text-color-active"]=t.itemTextColorActiveInverted,n["--n-item-text-color-child-active"]=t.itemTextColorChildActiveInverted,n["--n-item-text-color-child-active-hover"]=t.itemTextColorChildActiveInverted,n["--n-item-text-color-active-hover"]=t.itemTextColorActiveHoverInverted,n["--n-item-icon-color"]=t.itemIconColorInverted,n["--n-item-icon-color-hover"]=t.itemIconColorHoverInverted,n["--n-item-icon-color-active"]=t.itemIconColorActiveInverted,n["--n-item-icon-color-active-hover"]=t.itemIconColorActiveHoverInverted,n["--n-item-icon-color-child-active"]=t.itemIconColorChildActiveInverted,n["--n-item-icon-color-child-active-hover"]=t.itemIconColorChildActiveHoverInverted,n["--n-item-icon-color-collapsed"]=t.itemIconColorCollapsedInverted,n["--n-item-text-color-horizontal"]=t.itemTextColorHorizontalInverted,n["--n-item-text-color-hover-horizontal"]=t.itemTextColorHoverHorizontalInverted,n["--n-item-text-color-active-horizontal"]=t.itemTextColorActiveHorizontalInverted,n["--n-item-text-color-child-active-horizontal"]=t.itemTextColorChildActiveHorizontalInverted,n["--n-item-text-color-child-active-hover-horizontal"]=t.itemTextColorChildActiveHoverHorizontalInverted,n["--n-item-text-color-active-hover-horizontal"]=t.itemTextColorActiveHoverHorizontalInverted,n["--n-item-icon-color-horizontal"]=t.itemIconColorHorizontalInverted,n["--n-item-icon-color-hover-horizontal"]=t.itemIconColorHoverHorizontalInverted,n["--n-item-icon-color-active-horizontal"]=t.itemIconColorActiveHorizontalInverted,n["--n-item-icon-color-active-hover-horizontal"]=t.itemIconColorActiveHoverHorizontalInverted,n["--n-item-icon-color-child-active-horizontal"]=t.itemIconColorChildActiveHorizontalInverted,n["--n-item-icon-color-child-active-hover-horizontal"]=t.itemIconColorChildActiveHoverHorizontalInverted,n["--n-arrow-color"]=t.arrowColorInverted,n["--n-arrow-color-hover"]=t.arrowColorHoverInverted,n["--n-arrow-color-active"]=t.arrowColorActiveInverted,n["--n-arrow-color-active-hover"]=t.arrowColorActiveHoverInverted,n["--n-arrow-color-child-active"]=t.arrowColorChildActiveInverted,n["--n-arrow-color-child-active-hover"]=t.arrowColorChildActiveHoverInverted,n["--n-item-color-hover"]=t.itemColorHoverInverted,n["--n-item-color-active"]=t.itemColorActiveInverted,n["--n-item-color-active-hover"]=t.itemColorActiveHoverInverted,n["--n-item-color-active-collapsed"]=t.itemColorActiveCollapsedInverted):(n["--n-group-text-color"]=t.groupTextColor,n["--n-color"]=t.color,n["--n-item-text-color"]=t.itemTextColor,n["--n-item-text-color-hover"]=t.itemTextColorHover,n["--n-item-text-color-active"]=t.itemTextColorActive,n["--n-item-text-color-child-active"]=t.itemTextColorChildActive,n["--n-item-text-color-child-active-hover"]=t.itemTextColorChildActiveHover,n["--n-item-text-color-active-hover"]=t.itemTextColorActiveHover,n["--n-item-icon-color"]=t.itemIconColor,n["--n-item-icon-color-hover"]=t.itemIconColorHover,n["--n-item-icon-color-active"]=t.itemIconColorActive,n["--n-item-icon-color-active-hover"]=t.itemIconColorActiveHover,n["--n-item-icon-color-child-active"]=t.itemIconColorChildActive,n["--n-item-icon-color-child-active-hover"]=t.itemIconColorChildActiveHover,n["--n-item-icon-color-collapsed"]=t.itemIconColorCollapsed,n["--n-item-text-color-horizontal"]=t.itemTextColorHorizontal,n["--n-item-text-color-hover-horizontal"]=t.itemTextColorHoverHorizontal,n["--n-item-text-color-active-horizontal"]=t.itemTextColorActiveHorizontal,n["--n-item-text-color-child-active-horizontal"]=t.itemTextColorChildActiveHorizontal,n["--n-item-text-color-child-active-hover-horizontal"]=t.itemTextColorChildActiveHoverHorizontal,n["--n-item-text-color-active-hover-horizontal"]=t.itemTextColorActiveHoverHorizontal,n["--n-item-icon-color-horizontal"]=t.itemIconColorHorizontal,n["--n-item-icon-color-hover-horizontal"]=t.itemIconColorHoverHorizontal,n["--n-item-icon-color-active-horizontal"]=t.itemIconColorActiveHorizontal,n["--n-item-icon-color-active-hover-horizontal"]=t.itemIconColorActiveHoverHorizontal,n["--n-item-icon-color-child-active-horizontal"]=t.itemIconColorChildActiveHorizontal,n["--n-item-icon-color-child-active-hover-horizontal"]=t.itemIconColorChildActiveHoverHorizontal,n["--n-arrow-color"]=t.arrowColor,n["--n-arrow-color-hover"]=t.arrowColorHover,n["--n-arrow-color-active"]=t.arrowColorActive,n["--n-arrow-color-active-hover"]=t.arrowColorActiveHover,n["--n-arrow-color-child-active"]=t.arrowColorChildActive,n["--n-arrow-color-child-active-hover"]=t.arrowColorChildActiveHover,n["--n-item-color-hover"]=t.itemColorHover,n["--n-item-color-active"]=t.itemColorActive,n["--n-item-color-active-hover"]=t.itemColorActiveHover,n["--n-item-color-active-collapsed"]=t.itemColorActiveCollapsed),n}),w=r?ie("menu",C(()=>e.inverted?"a":"b"),z,e):void 0,K=ro(),V=$(null),ae=$(null);let _=!0;const ze=()=>{var m;_?_=!1:(m=V.value)===null||m===void 0||m.sync({showAllItemsBeforeCalculate:!0})};function Fe(){return document.getElementById(K)}const ee=$(-1);function Ve(m){ee.value=e.options.length-m}function je(m){m||(ee.value=-1)}const Ke=C(()=>{const m=ee.value;return{children:m===-1?[]:e.options.slice(m)}}),De=C(()=>{const{childrenField:m,disabledField:y,keyField:t}=e;return ce([Ke.value],{getIgnored(b){return he(b)},getChildren(b){return b[m]},getDisabled(b){return b[y]},getKey(b){var P;return(P=b[t])!==null&&P!==void 0?P:b.name}})}),Ue=C(()=>ce([{}]).treeNodes[0]);function Ge(){var m;if(ee.value===-1)return u(me,{root:!0,level:0,key:"__ellpisisGroupPlaceholder__",internalKey:"__ellpisisGroupPlaceholder__",title:"···",tmNode:Ue.value,domId:K,isEllipsisPlaceholder:!0});const y=De.value.treeNodes[0],t=R.value,b=!!(!((m=y.children)===null||m===void 0)&&m.some(P=>t.includes(P.key)));return u(me,{level:0,root:!0,key:"__ellpisisGroup__",internalKey:"__ellpisisGroup__",title:"···",virtualChildActive:b,tmNode:y,domId:K,rawNodes:y.rawNode.children||[],tmNodes:y.children||[],isEllipsisPlaceholder:!0})}return{mergedClsPrefix:o,controlledExpandedKeys:T,uncontrolledExpanededKeys:f,mergedExpandedKeys:p,uncontrolledValue:x,mergedValue:k,activePath:R,tmNodes:g,mergedTheme:i,mergedCollapsed:l,cssVars:r?void 0:z,themeClass:w?.themeClass,overflowRef:V,counterRef:ae,updateCounter:()=>{},onResize:ze,onUpdateOverflow:je,onUpdateCount:Ve,renderCounter:Ge,getCounter:Fe,onRender:w?.onRender,showOption:A,deriveResponsiveState:ze}},render(){const{mergedClsPrefix:e,mode:o,themeClass:r,onRender:i}=this;i?.();const a=()=>this.tmNodes.map(d=>ye(d,this.$props)),c=o==="horizontal"&&this.responsive,s=()=>u("div",no(this.$attrs,{role:o==="horizontal"?"menubar":"menu",class:[`${e}-menu`,r,`${e}-menu--${o}`,c&&`${e}-menu--responsive`,this.mergedCollapsed&&`${e}-menu--collapsed`],style:this.cssVars}),c?u(Co,{ref:"overflowRef",onUpdateOverflow:this.onUpdateOverflow,getCounter:this.getCounter,onUpdateCount:this.onUpdateCount,updateCounter:this.updateCounter,style:{width:"100%",display:"flex",overflow:"hidden"}},{default:a,counter:this.renderCounter}):a());return c?u(to,{onResize:this.onResize},{default:s}):s()}}),Xo={class:"nps-brand flex items-center gap-3 px-4 h-14"},Zo={class:"text-sm opacity-70 flex items-center gap-2"},Jo={class:"font-semibold nps-gradient-text"},Qo=B({__name:"MainLayout",setup(e){const{t:o,locale:r}=io(),i=go(),a=fo(),l=ao(),c=so();function s(p){const g={dashboard:'<path d="M3 13h8V3H3v10zm0 8h8v-6H3v6zm10 0h8V11h-8v10zm0-18v6h8V3h-8z"/>',clients:'<path d="M21 4H3a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h7v2H7v2h10v-2h-3v-2h7a1 1 0 0 0 1-1V5a1 1 0 0 0-1-1zm-1 12H4V6h16v10z"/>',hosts:'<path d="M12 2a10 10 0 1 0 10 10A10.011 10.011 0 0 0 12 2zm6.93 9h-3.96a14 14 0 0 0-1.21-5.06A8.014 8.014 0 0 1 18.93 11zM12 4c.81 0 1.97 2.04 2.45 5h-4.9C10.03 6.04 11.19 4 12 4zM4.07 11A8.014 8.014 0 0 1 9.24 5.94 14 14 0 0 0 8.03 11zm0 2h3.96a14 14 0 0 0 1.21 5.06A8.014 8.014 0 0 1 4.07 13zM12 20c-.81 0-1.97-2.04-2.45-5h4.9C13.97 17.96 12.81 20 12 20zm2.76-1.94A14 14 0 0 0 15.97 13h3.96a8.014 8.014 0 0 1-5.17 5.06z"/>',tcp:'<path d="M5 3h14v4H5zM5 10h14v4H5zM5 17h14v4H5z"/>',udp:'<path d="M3 13h2v-2H3v2zm4 0h2v-2H7v2zm4 0h2v-2h-2v2zm4 0h2v-2h-2v2zm4 0h2v-2h-2v2z"/>',http:'<path d="M3 5h18v2H3zM3 11h18v2H3zM3 17h18v2H3z"/>',socks5:'<path d="M12 2 4 6v6c0 5 3.5 9.7 8 10 4.5-.3 8-5 8-10V6l-8-4z"/>',secret:'<path d="M12 1 3 5v6c0 5.6 3.8 10.7 9 12 5.2-1.3 9-6.4 9-12V5l-9-4zm0 11h7c-.5 4-3.4 7.7-7 9V12H5V6l7-3v9z"/>',p2p:'<path d="M3 7h6v2H3zM3 11h10v2H3zM3 15h6v2H3zM15 7h6v2h-6zM15 11h6v2h-6zM15 15h6v2h-6z"/>',file:'<path d="M14 2H6a2 2 0 0 0-2 2v16c0 1.1.9 2 2 2h12a2 2 0 0 0 2-2V8l-6-6zm4 18H6V4h7v5h5v11z"/>',global:'<path d="M19.43 12.98c.04-.32.07-.65.07-.98s-.03-.66-.07-.98l2.11-1.65a.49.49 0 0 0 .12-.61l-2-3.46a.5.5 0 0 0-.61-.22l-2.49 1a7.03 7.03 0 0 0-1.7-.98l-.38-2.65A.49.49 0 0 0 14 2h-4a.49.49 0 0 0-.49.42l-.38 2.65c-.61.25-1.18.58-1.7.98l-2.49-1a.5.5 0 0 0-.61.22l-2 3.46a.49.49 0 0 0 .12.61l2.11 1.65c-.04.32-.07.65-.07.98s.03.66.07.98l-2.11 1.65a.49.49 0 0 0-.12.61l2 3.46c.14.24.43.34.69.22l2.49-1c.52.4 1.09.73 1.7.98l.38 2.65c.04.24.25.42.49.42h4c.24 0 .45-.18.49-.42l.38-2.65c.61-.25 1.18-.58 1.7-.98l2.49 1c.26.12.55.02.69-.22l2-3.46a.49.49 0 0 0-.12-.61l-2.11-1.65zM12 15.5a3.5 3.5 0 1 1 0-7 3.5 3.5 0 0 1 0 7z"/>',tokens:'<path d="M12 1 3 5v6c0 5.5 3.8 10.7 9 12 5.2-1.3 9-6.5 9-12V5l-9-4zm-1 6h2v6h-2V7zm0 8h2v2h-2v-2z"/>'};return()=>u(zo,{size:18},()=>u("svg",{viewBox:"0 0 24 24",width:18,height:18,fill:"currentColor",innerHTML:g[p]??""}))}const d=C(()=>[{label:o("nav.dashboard"),key:"dashboard",icon:s("dashboard")},{label:o("nav.clients"),key:"clients",icon:s("clients")},{type:"group",label:o("nav.tunnels"),key:"g-tunnels",children:[{label:o("nav.hosts"),key:"hosts",icon:s("hosts")},{label:o("nav.tcp"),key:"tunnels-tcp",icon:s("tcp")},{label:o("nav.udp"),key:"tunnels-udp",icon:s("udp")},{label:o("nav.http"),key:"tunnels-http",icon:s("http")},{label:o("nav.socks5"),key:"tunnels-socks5",icon:s("socks5")},{label:o("nav.secret"),key:"tunnels-secret",icon:s("secret")},{label:o("nav.p2p"),key:"tunnels-p2p",icon:s("p2p")},{label:o("nav.file"),key:"tunnels-file",icon:s("file")}]},{type:"group",label:o("nav.system"),key:"g-system",children:[{label:o("nav.global"),key:"global",icon:s("global")},{label:o("nav.tokens"),key:"tokens",icon:s("tokens")}]}]),x=C(()=>{const p=i.name;return p==="tunnels"?"tunnels-"+i.params.mode:p??""});function M(p){p.startsWith("tunnels-")?a.push({name:"tunnels",params:{mode:p.slice(8)}}):a.push({name:p})}const k=[{label:"简体中文",value:"zh-CN"},{label:"English",value:"en"}];function f(p){r.value=p,c.setLang(p)}const H=[{label:o("app.logout"),key:"logout"}];async function T(p){p==="logout"&&(await l.logout(),a.replace({name:"login"}))}return(p,g)=>(uo(),co(N(Se),{"has-sider":"",class:"h-screen"},{default:F(()=>[E(N($o),{bordered:"","collapse-mode":"width","collapsed-width":64,width:240,"show-trigger":"",collapsed:N(c).sidebarCollapsed,"onUpdate:collapsed":g[0]||(g[0]=R=>N(c).setSidebar(R))},{default:F(()=>[q("div",Xo,[g[2]||(g[2]=q("div",{class:"nps-brand-logo"},"N",-1)),vo(q("span",{class:"font-semibold tracking-wide nps-gradient-text text-[15px]"},oe(N(o)("app.title")),513),[[ho,!N(c).sidebarCollapsed]])]),E(N(Yo),{collapsed:N(c).sidebarCollapsed,"collapsed-width":64,"collapsed-icon-size":20,indent:16,options:d.value,value:x.value,"onUpdate:value":M},null,8,["collapsed","options","value"])]),_:1},8,["collapsed"]),E(N(Se),null,{default:F(()=>[E(N(_o),{bordered:"",class:"px-5 h-14 flex items-center justify-between nps-header"},{default:F(()=>[q("span",Zo,[g[3]||(g[3]=q("span",{class:"nps-status-dot"},null,-1)),te(" "+oe(N(o)("app.welcome"))+" ",1),q("span",Jo,oe(N(l).user?.username||"-"),1)]),E(N(Io),{align:"center",size:14},{default:F(()=>[E(N(So),{value:N(c).dark,"onUpdate:value":g[1]||(g[1]=R=>N(c).setDark(R))},{checked:F(()=>[...g[4]||(g[4]=[te("🌙",-1)])]),unchecked:F(()=>[...g[5]||(g[5]=[te("☀️",-1)])]),_:1},8,["value"]),E(N(yo),{value:N(r),options:k,size:"small",style:{width:"120px"},"onUpdate:value":f},null,8,["value"]),E(N(Ae),{options:H,onSelect:T},{default:F(()=>[E(N(mo),{size:"small",quaternary:""},{default:F(()=>[te(oe(N(l).user?.username||"-")+" ▾ ",1)]),_:1})]),_:1})]),_:1})]),_:1}),E(N(Ao),{class:"p-4"},{default:F(()=>[E(N(po))]),_:1})]),_:1})]),_:1}))}}),st=Ro(Qo,[["__scopeId","data-v-ff3d23bf"]]);export{st as default};
