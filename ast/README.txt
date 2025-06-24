See [[./ast.go]] for documentation.

To build the AST submodule and create the railroad diagram:

	go build && ./ast -diagram | railroad -o index.html -w && python -m http.server 8000
