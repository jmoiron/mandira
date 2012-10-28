# મંદિર Mandira

Mandira is a language agnostic logic-light templating system desigend to be usable on server and client side applications.  Mandira strives to be easy to learn, easy to implement, efficient to render, comfortable for designers, tollerable for developers, suitable for non-HTML documents, as flexible as necessary (but no more so), and consistent in style and rendering.

<!-- It is heavily influenced by [Mustache](http://mustache.github.com/mustache.5.html), [Tempo](http://tempojs.com/), [Django Templates](https://docs.djangoproject.com/en/dev/ref/templates/) and [Jinja2](http://jinja.pocoo.org/docs/). -->

Mandira templates should look familiar to users of other templates.  It is called "logic-light" because it allows some forms of conditional logic while avoiding the complexity of more sophisticated features.  It strives for both simplicity and ergonomics, but makes concessions to each where appropriate.

## Sample

A simple Mandira template:

```
Hello {{name}}
You have just won ${{value}}!
{{?if in_monaco}}
Well, ${{taxed_value}}, after taxes.
{{/if}}
```

Given the following context:

```
{
  "name": "Jason",
  "value": 10000,
  "taxed_value": 10000.0,
  "in_monaco": true
}
```

Will produce the following:

```
Hello Jason
You have just won $10000!
Well, $10000.0, after taxes.
```

## Language

Tags in Mandira are in `{{double-braces}}`, and the part inside the tag is called the tag **key**.

### Variables

**Variable** tags look like `{{name}}`.  This will attempt to look up the `name` key in the current context and render it.  If `name` isn't found, nothing is rendered.  Variables always are HTML-escaped.  To avoid escaping, use `{{{name}}}` with three braces instead.

Template:

```
{{first}}
{{last}}
{{title}}
{{{title}}}
```

Context:

```
{
    "first": "Sübü'ätäi",
    "title": "&lt;b>Örlög baghatur&lt;/b>"
}
```

Output:

```
Sübü'ätäi

&amp;lt;b&amp;gt;Örlög baghatur&amp;lt;/b&amp;gt;
&lt;b&gt;Örlög baghatur&lt;/b&gt;
```

### Filters

Variable tags can also take **filters**, which modify the output of variables.  Filters are applied using the `|` character (like unix pipes), can take additional arguments which must be in parentheses, and can be chained together from left to right.  `{{first|upper|js}}` will take the `first` variable, upper case it, and then escape it for javascript.  The actual `first` variable in the context remains unchanged;  it is only the output that changes.

Template:

```
{{first|upper}} {{last}}
titles: {{titles|join(", ")}}
family: {{family|join("|")}}
```

Context:

```
{
    "first": "Gaius"
    "last": "Julius Caesar",
    "titles": ["Consul", "Imperator", "Pontifex Maximus", "Dictator"]
}
```

Output:

```
GAIUS Julius Caesar
titles: Consul, Imperator, Pontifex Maximus, Dictator
family: 
```

The following filters are provided by default:

* `upper` - variable in upper case
* `lower` - variable in lower case
* `title` - variable in title case
* `len` - length of the variable
* `index` - the numeric index of a list or string
* `slice` - a slice of a list or string
* `format` - printf-style formatted string
* `date` - PHP style date formatted string
* `join` - each element of the list joined with a string, `", "` by default
* `divisibleby` - `true` if a number is divisible by an argument

More detailed information on their behavior and return values is available in the [implmentation specification]().

### Sections

Sections set off blocks of text to be rendered one or more times with an altered context, depending on the value of the key in the current context.  `{{#names}}` begins a "names" section, and `{{/names}}` ends it.  

If the key is a list, the block is rendered once per item in the list, with that item accessible by `{{.}}`.  If the list contains maps, the block rendered with each map's namespace as the context.  If the key is a map, the block is rendered once with that map's namespace as the context.  List sections additionally get the context variable `.index`, which is the zero-indexed iteration of the loop.

Template:

```
{{#generals}}
{{.index}}: {{.}}
{{/generals}}

{{#generallist}}
{{last|upper}}, {{first}}
{{/generallist}}

{{#napoleon}}
{{first}} {{last}}
b. {{born}}, d. {{died}}
{{/napoleon}}
```

Context:

```
{
    "generals": ["Alexander The Great", "Hannibal Barca", "Gaius Julius Caesar"],
    "generallist": [
        {"first": "Friedrich", "last": "der Große",
        {"first": "અર્જુન"}
        {"first": "家康", "last": "徳川"}
    ],
    "napoleon": {
        "first": "Napoléon", 
        "last": "Bonaparte", 
        "born": "15 August 1769", 
        "died": "5 May 1821"
    }
}
```

Output:

```
0: Alexander The Great
1: Hannibal Barca
2: Gaius Julius Caesar

DER GROßE, Friedrich
, અર્જુન
徳川, 家康

Napoléon Bonaparte
b. 15 August 1769, d. 5 May 1821
```

### Conditionals

Conditional logic is done via a section which looks like `{{?if condition}}...{{?else}}...{{/if}}`, where the `else` condition is optional.  The condition can be any variable + filter expression, or any number of these expressions combined with the following standard boolean operators:

* numeric operators `==`, `!=`, `<`, `>`, `<=`, `>=`
* logical operators `not`, `and`, `or`

Mandira allows literals in expressions for use in conditional logic and as arguments to filters.  The following shows examples of each data type, with the *falsey* value shown first:

* integers `0`, `1`, `213`
* floats: `0.0`, `1.4`, `3.14`
* strings: `""`,  `"hello, world!"`, `"મંદિર"`
* booleans: `false`, `true`

Putting it all together:

Template:

```
{{#generals}}
  {{?if born|slice(-2) == "BC" }}
    &lt;span class="ancient">{{name}}, b. {{born}}&lt;/span>
  {{?else}}
    &lt;span class="modern">{{name}}, b. {{born}}&lt;/span>
  {{/if}}
{{/generals}}
```

Context:

```
{
    "generals": [
        {"name": "Napoléon Bonaparte", "born": "15 August 1769"},
        {"name": "Gaius Julius Casear", "born": "July 100 BC"}
    ]
}
```

Output:

```
    &lt;span class="modern">Napoléon Bonaparte, b. 15 August 1769&lt;/span>
    &lt;span class="ancient">Gaius Julius Caesar, b. July 100 BC&lt;/span>
```

## Status

Currently, Mandira is only this loose specification.  I intend to get feedback on it and create a true specification, a reference implementation in [Go](http://golang.org) (maybe not in that order), and then live with it a while and see what works and what doesn't.

If you are interested in this language and the ideas behind it, read my [Logic-less Template Redux](http://jmoiron.net/blog/logicless-template-redux/) blog post or ping [@jmoiron](http://twitter/com/jmoiron) on twitter.  The name Mandira is sanskrit for "Temple", and the script used in the title is *Mandira* in [Gujarati](http://en.wikipedia.org/wiki/Gujarati_language).

