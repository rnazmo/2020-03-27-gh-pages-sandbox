

# Generate /contents/index.html .
.PHONY: gen
gen:
	go build -o ./gen/bin/a.out ./gen/main.go
	./gen/bin/a.out

.PHONY: push
push: gen
	git add .
	git commit -m "update"
	git push origin gh-pages

