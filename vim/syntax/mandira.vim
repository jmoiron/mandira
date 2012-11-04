" Mandira syntax
" Language:	Mandira
" Maintainer:	Jason Moiron <jmoiron@jmoiron.net>
" Screenshot:   
" Version:	1
" Last Change:  Nov 3 2012
" Remark:       
" Based un Juvenn Woo <machese@gmail.com>'s mustache highlighting and
" Armin Ronacher's jinja.vim.
" References:	
" TODO: Feedback is welcomed.


" Read the HTML syntax to start with
if version < 600
  so <sfile>:p:h/html.vim
else
  runtime! syntax/html.vim
  unlet b:current_syntax
endif

if version < 600
  syntax clear
elseif exists("b:current_syntax")
  finish
endif

" Standard HiLink will not work with included syntax files
if version < 508
  command! -nargs=+ HtmlHiLink hi link <args>
else
  command! -nargs=+ HtmlHiLink hi def link <args>
endif

syntax match mandiraError '}}}\?'
syntax match mandiraInsideError '{{[{#<>=!\/]\?' containedin=@mandiraInside
syntax region mandiraVariable matchgroup=mandiraMarker start=/{{/ end=/}}/ containedin=@htmlMustacheContainer 
syntax region mandiraVariableUnescape matchgroup=mandiraMarker start=/{{{/ end=/}}}/ containedin=@htmlMustacheContainer
syntax region mandiraSection matchgroup=mandiraMarker start='{{[#/]' end=/}}/ containedin=@htmlMustacheContainer
" syntax region mandiraPartial matchgroup=mandiraMarker start=/{{[<>]/ end=/}}/
syntax region mandiraConditional matchgroup=mandiraMarker start='{{[?]' end=/}}/ containedin=@htmlMustacheContainer
" syntax region mandiraMarkerSet matchgroup=mandiraMarker start=/{{=/ end=/=}}/
syntax region mandiraComment start=/{{!/ end=/}}/ contains=Todo containedin=htmlHead
syntax region mandiraString containedin=mandiraConditional,mandiraSection,mandiraVariable,mandiraVariableUnescape contained start=/"/ skip=/\\"/ end=/"/
syntax keyword mandiraOperator containedin=mandiraConditional,mandiraSection contained and if else not or
syntax match mandiraOperator "|" containedin=mandiraConditional,mandiraSection,mandiraVariable contained nextgroup=mandiraFilter
syntax match mandiraFilter contained skipwhite /[a-zA-Z_][a-zA-Z0-9_]*/


" Clustering
syntax cluster mandiraInside add=mandiraVariable,mandiraVariableUnescape,mandiraSection,mandiraPartial,mandiraMarkerSet
syntax cluster htmlMustacheContainer add=htmlHead,htmlTitle,htmlString,htmlH1,htmlH2,htmlH3,htmlH4,htmlH5,htmlH6


" Hilighting
" Inside hilighted as Number, which is rarely used in html
" you might like change it to Function or Identifier
HtmlHiLink mandiraOperator Keyword
HtmlHiLink mandiraString Special
HtmlHiLink mandiraConditional Number
HtmlHiLink mandiraFilter Function

HtmlHiLink mandiraVariable Number
HtmlHiLink mandiraVariableUnescape Number
HtmlHiLink mandiraPartial Number
HtmlHiLink mandiraSection Number
HtmlHiLink mandiraMarkerSet Number

HtmlHiLink mandiraComment Comment
HtmlHiLink mandiraMarker Identifier
HtmlHiLink mandiraError Error
HtmlHiLink mandiraInsideError Error

syn region mandiraScriptTemplate start=+<script [^>]*type *=[^>]*text/mandira[^>]*>+
\                       end=+</script>+me=s-1 keepend
\                       contains=mandiraError,mandiraInsideError,mandiraVariable,mandiraVariableUnescape,mandiraSection,mandiraPartial,mandiraMarkerSet,mandiraComment,htmlHead,htmlTitle,htmlString,htmlH1,htmlH2,htmlH3,htmlH4,htmlH5,htmlH6,htmlTag,htmlEndTag,htmlTagName,htmlSpecialChar,htmlLink

let b:current_syntax = "mandira"
delcommand HtmlHiLink
