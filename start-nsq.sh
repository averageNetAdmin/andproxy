nsqlookupd && 
nsqd --lookupd-tcp-address=127.0.0.1:4160 &&
nsqadmin --lookupd-http-address=127.0.0.1:4161
# nsq_to_file --topic=test --output-dir=/tmp --lookupd-http-address=127.0.0.1:4161 &&