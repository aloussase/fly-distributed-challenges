all:
	@echo Tasks: unique-ids, broadcast

unique-ids: cmd/unique-ids
	go build ./$<
	../maelstrom/maelstrom test -w unique-ids --bin ./$@ --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

broadcast: cmd/broadcast
	go build ./$<
	../maelstrom/maelstrom test -w broadcast --bin ./$@ --node-count 1 --time-limit 20 --rate 10

clean:
	rm -rf echo unique-ids broadcast

.PHONY: clean