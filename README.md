# મંદિર Mandira

Mandira is a language agnostic logic-light templating system desigend to be usable on server and client side applications.  Mandira strives to be easy to learn, easy to implement, efficient to render, comfortable for designers, tollerable for developers, suitable for non-HTML documents, as flexible as necessary (but no more so), and consistent in style and rendering.

It is heavily influenced by [Mustache](http://mustache.github.com/mustache.5.html), [Tempo](http://tempojs.com/), [Django Templates](https://docs.djangoproject.com/en/dev/ref/templates/) and [Jinja2](http://jinja.pocoo.org/docs/)

Mandira templates should look familiar to users of other templates.  It is called "logic-light" because it allows some forms of conditional logic while avoiding the complexity of more sophisticated features.  It strives for both simplicity and ergonomics, but makes concessions to each where appropriate.

## Sample

A simple Mandira template:

```
{{#greetings}}
  {{?if greeting|len > 4}}
    {{greeting}}
  {{?else}}
    {{greeting}}, {{name}} ({{age}})
  {{/if}}
{{/greetings}}
```

Given the following context:

```
{
  "name": "Jason",
  "greetings": [
    {"greeting": "Hello"},
    {"greeting": "Hi"}
  ],
  "age": 30
}
```

Will produce the following:

```
Hello
Hi, Jason (30)
```


## Status

Currently, Mandira has [a specification](http://jmoiron.github.com/mandira) and an **alpha quality** reference implementation in [Go](http://golang.org).  It is currently the template language in use on [my personal website](http://jmoiron.net).  I intend to get feedback on it and live with it a while and see what works and what doesn't.

If you are interested in this language and the ideas behind it, read my [Logic-less Template Redux](http://jmoiron.net/blog/logicless-template-redux/) blog post or ping [@jmoiron](http://twitter/com/jmoiron) on twitter.  The name Mandira is sanskrit for "Temple", and the script used in the title is *Mandira* in [Gujarati](http://en.wikipedia.org/wiki/Gujarati_language).

