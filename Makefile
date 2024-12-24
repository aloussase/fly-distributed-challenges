all:
	@echo Tasks: unique-ids, broadcast

unique-ids: cmd/unique-ids
	go build ./$<
	../maelstrom/maelstrom test -w unique-ids --bin ./$@ --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

broadcast: cmd/broadcast
	go build ./$<
	../maelstrom/maelstrom test -w broadcast --bin ./$@ --node-count 1 --time-limit 20 --rate 10

multibroadcast: cmd/multibroadcast
	go build ./$<
	../maelstrom/maelstrom test -w broadcast --bin ./$@ --node-count 5 --time-limit 20 --rate 10

fault-tolerant-multibroadcast: cmd/fault-tolerant-multibroadcast
	go build ./$<
	../maelstrom/maelstrom test -w broadcast --bin ./$@ --node-count 5 --time-limit 20 --rate 10 --nemesis partition

clean:
	rm -rf echo unique-ids broadcast multibroadcast fault-tolerant-multibroadcast

.PHONY: clean