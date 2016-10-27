package test_juno

import "strconv"

//http://redis.io/topics/protocol#resp-protocol-description
const (
	prf_err  = "-"
	prf_str  = "+"
	prf_int  = ":"
	prf_bulk = "$"
	prf_arr  = "*"
)

//http://redis.io/topics/protocol#resp-errors
func format_err(errType string, descr string) string {
	return prf_err + errType + "\r\n" + descr + "\r\n"
}

//http://redis.io/topics/protocol#resp-simple-strings
func format_str(str string) string {
	return prf_str + str + "\r\n"
}

//http://redis.io/topics/protocol#resp-integers
func format_int(val int64) string {
	return prf_int + strconv.FormatInt(val, 10) + "\r\n"
}

//http://redis.io/topics/protocol#resp-bulk-strings
func format_bulk(str string) string {
	return prf_bulk + strconv.FormatInt(int64(len(str)), 10) + "\r\n" + str + "\r\n"
}

//http://redis.io/topics/protocol#resp-arrays
func format_array(length int) string {
	return prf_arr + strconv.FormatInt(int64(length), 10) + "\r\n"
}

//http://redis.io/topics/protocol#nil-reply
func format_nill() string {
	return prf_bulk + "-1\r\n"
}

func format_standart_err(descr string) string {
	return format_err("ERR", descr)
}

func format_ok() string {
	return format_str("OK")
}
