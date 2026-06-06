See [[./doc.go]] for documentation.

To build the AST submodule and create a railroad diagram:

	go build && ./ast -diagram | railroad -o index.html -w && python -m http.server 8000

To-do:

- use Walk for evaluation
- determine recursion/depth limiters
- isResolvable test for nested expressions
