.PHONY: always clean dist

.everything: always
	export SUM=$$(find . -type f \( ! -path "./$@" ! -path "./Makefile" ! -path "./.git/*" \) -print0 | sort -z | xargs -0 sha1sum | sha1sum); \
	echo "$$SUM" | cmp -s - $@ || echo "$$SUM" > $@

server_image: .everything
	touch server_image

dist: server_image

clean:
	rm .everything
