.PHONY: setuptests


setuptests:
	aws s3 cp testdata.txt s3://s3-check-role-2025/testdata.txt
	aws s3 ls --recursive s3://s3-check-role-2025/

.PHONY: testread
testread:
	./test_read_s3.sh

.PHONY: testwritefails
testwritefails:
	./test_write_s3_fails.sh

.PHONY: test
test: setuptests testread testwritefails