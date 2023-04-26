maelstrom_deps:
	sudo apt-get update
	sudo apt-get install graphviz gnuplot

maelstrom_fetch:
	wget -P ./third-party https://github.com/jepsen-io/maelstrom/releases/download/v0.2.3/maelstrom.tar.bz2 && bzip2 -d ./third-party/maelstrom.tar.bz2 && tar -xvf ./third-party/maelstrom.tar -C ./third-party && rm ./third-party/maelstrom.tar

maelstrom_serve:
	./third-party/maelstrom/maelstrom serve