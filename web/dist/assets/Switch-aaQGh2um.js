import{Q as ge,c as ve,aw as pe,j as me,ax as we,n as X,C as i,ay as Y,q as H,p as s,X as q,d as xe,s as J,az as N,t as b,v as C,f as _,x as l,G as k,y as r,R as G,E as Q,z as ye,ao as ke,b as E,K as Se,D,a9 as O,aA as Be,M as Ce,aB as _e,O as S,aC as U,aD as d,I as ze}from"./index-BLS79P1g.js";import{u as $e}from"./_plugin-vue_export-helper-DRL5z3rr.js";function Te(){const e=ve(pe,null);return e===null&&ge("use-dialog","No outer <n-dialog-provider /> founded."),e}var Re={buttonHeightSmall:"14px",buttonHeightMedium:"18px",buttonHeightLarge:"22px",buttonWidthSmall:"14px",buttonWidthMedium:"18px",buttonWidthLarge:"22px",buttonWidthPressedSmall:"20px",buttonWidthPressedMedium:"24px",buttonWidthPressedLarge:"28px",railHeightSmall:"18px",railHeightMedium:"22px",railHeightLarge:"26px",railWidthSmall:"32px",railWidthMedium:"40px",railWidthLarge:"48px"};function Ve(e){const{primaryColor:c,opacityDisabled:f,borderRadius:o,textColor3:g}=e;return{...Re,iconColor:g,textColor:"white",loadingColor:c,opacityDisabled:f,railColor:"rgba(0, 0, 0, .14)",railColorActive:c,buttonBoxShadow:"0 1px 4px 0 rgba(0, 0, 0, 0.3), inset 0 0 1px 0 rgba(0, 0, 0, 0.05)",buttonColor:"#FFF",railBorderRadiusSmall:o,railBorderRadiusMedium:o,railBorderRadiusLarge:o,buttonBorderRadiusSmall:o,buttonBorderRadiusMedium:o,buttonBorderRadiusLarge:o,boxShadowFocus:`0 0 0 2px ${we(c,{alpha:.2})}`}}const Fe={common:me,self:Ve};var We=X("switch",`
 height: var(--n-height);
 min-width: var(--n-width);
 vertical-align: middle;
 user-select: none;
 -webkit-user-select: none;
 display: inline-flex;
 outline: none;
 justify-content: center;
 align-items: center;
`,[i("children-placeholder",`
 height: var(--n-rail-height);
 display: flex;
 flex-direction: column;
 overflow: hidden;
 pointer-events: none;
 visibility: hidden;
 `),i("rail-placeholder",`
 display: flex;
 flex-wrap: none;
 `),i("button-placeholder",`
 width: calc(1.75 * var(--n-rail-height));
 height: var(--n-rail-height);
 `),X("base-loading",`
 position: absolute;
 top: 50%;
 left: 50%;
 transform: translateX(-50%) translateY(-50%);
 font-size: calc(var(--n-button-width) - 4px);
 color: var(--n-loading-color);
 transition: color .3s var(--n-bezier);
 `,[Y({left:"50%",top:"50%",originalTransform:"translateX(-50%) translateY(-50%)"})]),i("checked, unchecked",`
 transition: color .3s var(--n-bezier);
 color: var(--n-text-color);
 box-sizing: border-box;
 position: absolute;
 white-space: nowrap;
 top: 0;
 bottom: 0;
 display: flex;
 align-items: center;
 line-height: 1;
 `),i("checked",`
 right: 0;
 padding-right: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),i("unchecked",`
 left: 0;
 justify-content: flex-end;
 padding-left: calc(1.25 * var(--n-rail-height) - var(--n-offset));
 `),H("&:focus",[i("rail",`
 box-shadow: var(--n-box-shadow-focus);
 `)]),s("round",[i("rail","border-radius: calc(var(--n-rail-height) / 2);",[i("button","border-radius: calc(var(--n-button-height) / 2);")])]),q("disabled",[q("icon",[s("rubber-band",[s("pressed",[i("rail",[i("button","max-width: var(--n-button-width-pressed);")])]),i("rail",[H("&:active",[i("button","max-width: var(--n-button-width-pressed);")])]),s("active",[s("pressed",[i("rail",[i("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])]),i("rail",[H("&:active",[i("button","left: calc(100% - var(--n-offset) - var(--n-button-width-pressed));")])])])])])]),s("active",[i("rail",[i("button","left: calc(100% - var(--n-button-width) - var(--n-offset))")])]),i("rail",`
 overflow: hidden;
 height: var(--n-rail-height);
 min-width: var(--n-rail-width);
 border-radius: var(--n-rail-border-radius);
 cursor: pointer;
 position: relative;
 transition:
 opacity .3s var(--n-bezier),
 background .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 background-color: var(--n-rail-color);
 `,[i("button-icon",`
 color: var(--n-icon-color);
 transition: color .3s var(--n-bezier);
 font-size: calc(var(--n-button-height) - 4px);
 position: absolute;
 left: 0;
 right: 0;
 top: 0;
 bottom: 0;
 display: flex;
 justify-content: center;
 align-items: center;
 line-height: 1;
 `,[Y()]),i("button",`
 align-items: center; 
 top: var(--n-offset);
 left: var(--n-offset);
 height: var(--n-button-height);
 width: var(--n-button-width-pressed);
 max-width: var(--n-button-width);
 border-radius: var(--n-button-border-radius);
 background-color: var(--n-button-color);
 box-shadow: var(--n-button-box-shadow);
 box-sizing: border-box;
 cursor: inherit;
 content: "";
 position: absolute;
 transition:
 background-color .3s var(--n-bezier),
 left .3s var(--n-bezier),
 opacity .3s var(--n-bezier),
 max-width .3s var(--n-bezier),
 box-shadow .3s var(--n-bezier);
 `)]),s("active",[i("rail","background-color: var(--n-rail-color-active);")]),s("loading",[i("rail",`
 cursor: wait;
 `)]),s("disabled",[i("rail",`
 cursor: not-allowed;
 opacity: .5;
 `)])]);const Me=["aria-checked","tabindex","onClick","onFocus","onBlur","onKeyup","onKeydown"],Pe={...J.props,size:String,value:{type:[String,Number,Boolean],default:void 0},loading:Boolean,defaultValue:{type:[String,Number,Boolean],default:!1},disabled:{type:Boolean,default:void 0},round:{type:Boolean,default:!0},"onUpdate:value":[Function,Array],onUpdateValue:[Function,Array],checkedValue:{type:[String,Number,Boolean],default:!0},uncheckedValue:{type:[String,Number,Boolean],default:!1},railStyle:Function,rubberBand:{type:Boolean,default:!0},spinProps:Object,onChange:[Function,Array]};let F;var Le=xe({name:"Switch",props:Pe,slots:Object,setup(e){F===void 0&&(typeof CSS<"u"?typeof CSS.supports<"u"?F=CSS.supports("width","max(1px)"):F=!1:F=!0);const{mergedClsPrefixRef:c,inlineThemeDisabled:f,mergedComponentPropsRef:o}=ye(e),g=J("Switch","-switch",We,Fe,e,c),v=ke(e,{mergedSize(t){var x,y;if(e.size!==void 0)return e.size;if(t)return t.mergedSize.value;const w=(y=(x=o==null?void 0:o.value)==null?void 0:x.Switch)==null?void 0:y.size;return w||"medium"}}),{mergedSizeRef:z,mergedDisabledRef:p}=v,$=E(e.defaultValue),W=ze(e,"value"),m=$e(W,$),M=D(()=>m.value===e.checkedValue),a=E(!1),n=E(!1),R=D(()=>{const{railStyle:t}=e;if(t)return t({focused:n.value,checked:M.value})});function P(t){const{"onUpdate:value":w,onChange:x,onUpdateValue:y}=e,{nTriggerFormInput:K,nTriggerFormChange:T}=v;w&&O(w,t),y&&O(y,t),x&&O(x,t),$.value=t,K(),T()}function Z(){const{nTriggerFormFocus:t}=v;t()}function ee(){const{nTriggerFormBlur:t}=v;t()}function te(){e.loading||p.value||(m.value!==e.checkedValue?P(e.checkedValue):P(e.uncheckedValue))}function ie(){n.value=!0,Z()}function ae(){n.value=!1,ee(),a.value=!1}function oe(t){e.loading||p.value||t.key===" "&&(m.value!==e.checkedValue?P(e.checkedValue):P(e.uncheckedValue),a.value=!1)}function re(t){e.loading||p.value||t.key===" "&&(t.preventDefault(),a.value=!0)}const I=D(()=>{const{value:t}=z,{self:{opacityDisabled:w,railColor:x,railColorActive:y,buttonBoxShadow:K,buttonColor:T,boxShadowFocus:ne,loadingColor:le,textColor:se,iconColor:de,[S("buttonHeight",t)]:u,[S("buttonWidth",t)]:ce,[S("buttonWidthPressed",t)]:ue,[S("railHeight",t)]:h,[S("railWidth",t)]:V,[S("railBorderRadius",t)]:he,[S("buttonBorderRadius",t)]:be},common:{cubicBezierEaseInOut:fe}}=g.value;let L,j,A;return F?(L=`calc((${h} - ${u}) / 2)`,j=`max(${h}, ${u})`,A=`max(${V}, calc(${V} + ${u} - ${h}))`):(L=U((d(h)-d(u))/2),j=U(Math.max(d(h),d(u))),A=d(h)>d(u)?V:U(d(V)+d(u)-d(h))),{"--n-bezier":fe,"--n-button-border-radius":be,"--n-button-box-shadow":K,"--n-button-color":T,"--n-button-width":ce,"--n-button-width-pressed":ue,"--n-button-height":u,"--n-height":j,"--n-offset":L,"--n-opacity-disabled":w,"--n-rail-border-radius":he,"--n-rail-color":x,"--n-rail-color-active":y,"--n-rail-height":h,"--n-rail-width":V,"--n-width":A,"--n-box-shadow-focus":ne,"--n-loading-color":le,"--n-text-color":se,"--n-icon-color":de}}),B=f?Se("switch",D(()=>z.value[0]),I,e):void 0;return{handleClick:te,handleBlur:ae,handleFocus:ie,handleKeyup:oe,handleKeydown:re,mergedRailStyle:R,pressed:a,mergedClsPrefix:c,mergedValue:m,checked:M,mergedDisabled:p,cssVars:f?void 0:I,themeClass:B==null?void 0:B.themeClass,onRender:B==null?void 0:B.onRender}},render(){const{mergedClsPrefix:e,mergedDisabled:c,checked:f,mergedRailStyle:o,onRender:g,$slots:v}=this;g==null||g();const{checked:z,unchecked:p,icon:$,"checked-icon":W,"unchecked-icon":m}=v,M=!(N($)&&N(W)&&N(m));return b(),C("div",{role:"switch","aria-checked":f,class:r([`${e}-switch`,this.themeClass,M&&`${e}-switch--icon`,f&&`${e}-switch--active`,c&&`${e}-switch--disabled`,this.round&&`${e}-switch--round`,this.loading&&`${e}-switch--loading`,this.pressed&&`${e}-switch--pressed`,this.rubberBand&&`${e}-switch--rubber-band`]),tabindex:this.mergedDisabled?void 0:0,style:Q(this.cssVars),onClick:this.handleClick,onFocus:this.handleFocus,onBlur:this.handleBlur,onKeyup:this.handleKeyup,onKeydown:this.handleKeydown},[_("div",{class:r(`${e}-switch__rail`),"aria-hidden":"true",style:Q(o)},[l(()=>k(z,a=>k(p,n=>a||n?(b(),C("div",{key:4,"aria-hidden":!0,class:r(`${e}-switch__children-placeholder`)},[_("div",{class:r(`${e}-switch__rail-placeholder`)},[_("div",{class:r(`${e}-switch__button-placeholder`)},null,2),l(()=>a)],2),_("div",{class:r(`${e}-switch__rail-placeholder`)},[_("div",{class:r(`${e}-switch__button-placeholder`)},null,2),l(()=>n)],2)],2)):null))),_("div",{class:r(`${e}-switch__button`)},[l(()=>k($,a=>k(W,n=>k(m,R=>(b(),G(_e,null,{default:()=>this.loading?(b(),G(Be,Ce({key:"loading",clsPrefix:e,strokeWidth:20},this.spinProps),null,16,["clsPrefix"])):this.checked&&(n||a)?(b(),C("div",{class:r(`${e}-switch__button-icon`),key:n?"checked-icon":"icon"},[l(()=>n||a)],2)):!this.checked&&(R||a)?(b(),C("div",{class:r(`${e}-switch__button-icon`),key:R?"unchecked-icon":"icon"},[l(()=>R||a)],2)):null},1024)))))),l(()=>k(z,a=>a&&(b(),C("div",{key:"checked",class:r(`${e}-switch__checked`)},[l(()=>a)],2)))),l(()=>k(p,a=>a&&(b(),C("div",{key:"unchecked",class:r(`${e}-switch__unchecked`)},[l(()=>a)],2))))],2)],6)],46,Me)}});export{Le as S,Te as u};
