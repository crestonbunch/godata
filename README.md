Godata
======

This is an implementation of OData in Go. It is capable of parsing an OData
request, and passing it to a provider to create a response.

Providers are custom-made components which convert a GoData request into a
GoData response. Providers typically connect Godata to a database. For example
a GoData request might be converted into the corresponding SQL statement(s) to
fetch the desired data, then package the result into a response.
