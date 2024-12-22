unique-ids: cmd/unique-ids
	go build ./$<
	../maelstrom/maelstrom test -w unique-ids --bin ./$@ --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

clean:
	rm -rf echo unique-ids

.PHONY: clean