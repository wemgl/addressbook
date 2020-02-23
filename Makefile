.PHONY: proto
proto:
	protoc -I proto/ --go_out=. proto/addressbookpb/*.proto
