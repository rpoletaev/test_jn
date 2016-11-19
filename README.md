# test_jn
redis like cache

supporting commands:

* PASS  - authentication

* GET - get value 
* SET - set value
* DEL - delete key
* KEYS  - list of all keys
* EXPIRE - setup expiration seconds
* TTL - returns ttl for key

* RPUSH - push to end list
* LPUSH - push to start list
* RPOP  - get end item of list
* LPOP  - get start item of list
* LINSERT - insert value on index
* LINDEX - get value of list by index
* LREM - remove list
* LINSAFT - insert to list after index
* LLEN  - returns length of list

* CMDS  - returns list of all commands
--------------------------------------
### instalation
go get github.com/rpoletaev/test_jn
