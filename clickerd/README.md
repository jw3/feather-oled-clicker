clickerd
===

clicker daemon that interfaces a clicker with an endpoint

### items

defined in yaml, selectable things that have a title, uri, and body. upon selection the body is sent to the uri.

- body is any string
- display limits are set based on font size
- uri is optional, will use default uri if not set

#### example

```
items:
- id: item1
  title: human readable
  uri: http://fooo.barr
  body: {"what":"embedded-json"}
```
