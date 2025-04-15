all: uninterruptible

uninterruptible-build: uninterruptible/uninterruptible
	gcc -o uninterruptible/uninterruptible uninterruptible/uninterruptible.c

uninterruptible: uninterruptible-build
	./uninterruptible/uninterruptible
