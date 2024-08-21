go build && ./ast -diagram | railroad -o index.html -w && python -m http.server 8000
