# મંદિર Mandira

Mandira is a language agnostic logic-light templating system desigend to be usable on server and client side applications.  Mandira strives to be easy to learn, easy to implement, efficient to render, comfortable for designers, tollerable for developers, suitable for non-HTML documents, as flexible as necessary (but no more so), and consistent in style and rendering.

It is heavily influenced by [Mustache](http://mustache.github.com/mustache.5.html), [Tempo](http://tempojs.com/), [Django Templates](https://docs.djangoproject.com/en/dev/ref/templates/) and [Jinja2](http://jinja.pocoo.org/docs/)

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

## Status

Currently, Mandira exists only as [a loose specification](http://jmoiron.github.com/mandira).  I intend to get feedback on it and create a true specification, a reference implementation in [Go](http://golang.org) (maybe not in that order), and then live with it a while and see what works and what doesn't.

If you are interested in this language and the ideas behind it, read my [Logic-less Template Redux](http://jmoiron.net/blog/logicless-template-redux/) blog post or ping [@jmoiron](http://twitter/com/jmoiron) on twitter.  The name Mandira is sanskrit for "Temple", and the script used in the title is *Mandira* in [Gujarati](http://en.wikipedia.org/wiki/Gujarati_language).

