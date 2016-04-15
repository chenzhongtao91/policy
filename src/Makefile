POLICY_EXEC_FILE = ./bin/policy

.PHONY: all clean

all: $(POLICY_EXEC_FILE)

FLAGS = -tags "libdm_no_deferred_remove"
ifeq ($(STATIC_LINK), 1)
    FLAGS = -a -tags "netgo libdm_no_deferred_remove" \
	    -ldflags "-linkmode external -extldflags -static" \
	    --installsuffix netgo
endif

$(POLICY_EXEC_FILE): ./main.go
	
	go build -v 

clean:
	rm -f $(POLICY_EXEC_FILE)

install:
	cp $(POLICY_EXEC_FILE) /host/bin

test:
	go test -tags "libdm_no_deferred_remove" ./...
