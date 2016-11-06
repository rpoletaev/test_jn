package resp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

//http://redis.io/topics/protocol#resp-protocol-description
const (
	prf_err  = '-'
	prf_str  = '+'
	prf_int  = ':'
	prf_bulk = '$'
	prf_arr  = '*'
)

var WrongCmdCommand error = errors.New("Wrong command format")

//http://redis.io/topics/protocol#resp-errors
func format_err(errType string, descr string) string {
	return string(prf_err) + errType + " " + descr + "\r\n"
}

//http://redis.io/topics/protocol#resp-simple-strings
func format_str(str string) string {
	return string(prf_str) + str + "\r\n"
}

//http://redis.io/topics/protocol#resp-integers
func format_int(val int64) string {
	return string(prf_int) + strconv.FormatInt(val, 10) + "\r\n"
}

//http://redis.io/topics/protocol#resp-bulk-strings
func format_bulk_string(str ...string) string {
	fmt.Printf("%q\n", str)
	fmt.Println(string(prf_bulk) + strconv.FormatInt(int64(len(str)), 10) + "\r\n" + strings.Join(str, "\r\n"))
	return string(prf_bulk) + strconv.FormatInt(int64(len(str)), 10) + "\r\n" + strings.Join(str, "\r\n") + "\r\n"
}

//http://redis.io/topics/protocol#resp-arrays
func format_array(val []interface{}, length int) string {
	return string(prf_arr) + strconv.FormatInt(int64(length), 10) + "\r\n"
}

//http://redis.io/topics/protocol#nil-reply
func format_nill() string {
	return string(prf_bulk) + "-1\r\n"
}

func format_standart_err(descr string) string {
	return format_err("ERR", descr)
}

func format_ok() string {
	return format_str("OK")
}

type RespCommand struct {
	Name string
	prs  []interface{}
}

func isArray(args []string) (yes bool, tail []string, itemsCount int64, err error) {
	tail = args
	if yes = tail[0][0] == prf_arr; !yes {
		return yes, tail, itemsCount, nil
	}

	itemsCount, err = strconv.ParseInt(tail[0][1:], 10, 64)
	if err != nil {
		return yes, tail, itemsCount, err
	}

	tail = tail[1:]
	return yes, tail, itemsCount, nil
}

func ParseIfInt(args []string) (yes bool, result int64, tail []string, err error) {
	tail = args
	if yes = tail[0][0] == prf_int; !yes {
		return yes, result, tail, nil
	}

	if len(tail[0]) < 2 {
		return yes, result, tail, errors.New("WrongRespIntFormat")
	}

	result, err = strconv.ParseInt(tail[0][1:], 10, 64)
	if err != nil {
		return yes, result, tail, err
	}

	tail = tail[1:]
	return yes, result, tail, err
}

func ParseIfString(args []string) (yes bool, result string, tail []string, err error) {
	tail = args
	if yes = tail[0][0] == prf_str; !yes {
		return yes, result, tail, nil
	}

	if len(tail[0]) < 2 {
		return yes, result, tail, errors.New("WrongRespStringFormat")
	}

	result = tail[0][1:]
	tail = tail[1:]
	return yes, result, tail, nil
}

func ParseIfError(args []string) (yes bool, result error, tail []string, err error) {
	tail = args
	if yes = tail[0][0] == prf_err; !yes {
		return yes, result, tail, nil
	}

	if len(tail[0]) < 2 {
		return yes, result, tail, errors.New("WrongRespStringFormat")
	}

	result = errors.New(tail[0][1:])
	tail = tail[1:]
	return yes, result, tail, nil
}

func ParseIfBulkString(args []string) (yes bool, result string, tail []string, err error) {
	tail = args
	if yes = tail[0][0] == prf_bulk; !yes {
		return yes, result, tail, nil
	}

	if len(tail[0]) < 2 {
		return yes, result, tail, errors.New("WrongBulkStringFormat")
	}

	var bulkLen int64
	bulkLen, err = strconv.ParseInt(tail[0][1:], 10, 64)
	if err != nil {
		return yes, result, tail, errors.New("WrongBulkStringFormat")
	}

	if utf8.RuneCountInString(tail[1]) < int(bulkLen) {
		return yes, result, tail, errors.New("WrongBulkStringFormat")
	}
	result = string([]rune(tail[1])[0:bulkLen])
	tail = tail[2:]
	return yes, result, tail, nil
}

func ParseArray(args []string, len int64) (result []interface{}, tail []string, err error) {
	//*2\r\n
	//$4\r\n
	//LLEN\r\n
	//$6\r\n
	//mylist\r\n
	tail = args
	result = make([]interface{}, len)
	for i := 0; i < int(len); i++ {
		fmt.Printf("qurrent tail %q\n", tail)
		var ok bool
		var res interface{}
		//Check Integer
		ok, res, tail, err = ParseIfInt(tail)
		if ok && err != nil {
			return result, tail, err
		}

		if ok {
			result[i] = res
			continue
		}

		//Check simple string
		ok, res, tail, err = ParseIfString(tail)
		if ok && err != nil {
			return result, tail, err
		}

		if ok {
			result[i] = res
			continue
		}

		//Check error
		//Check simple string
		ok, res, tail, err = ParseIfError(tail)
		if ok && err != nil {
			return result, tail, err
		}

		if ok {
			result[i] = res
			continue
		}

		//Check bulk string
		ok, res, tail, err = ParseIfBulkString(tail)
		if ok && err != nil {
			return result, tail, err
		}

		if ok {
			result[i] = res
			continue
		}

		var arLen int64
		ok, tail, arLen, err = isArray(tail)
		if ok && err != nil {
			return result, tail, err
		}

		if ok {
			res, tail, err = ParseArray(tail, arLen)
			if err != nil {
				return result, tail, err
			}

			result[i] = res
			continue
		}
		//Must be ureacheble
		return result, tail, errors.New("RESP parsing error. Uncknown RESP type")
	}
	return result, tail, nil
}

func ParseRespString(src string) (result []interface{}, err error) {
	tail := strings.Split(src, "\r\n")
	tail = tail[:len(tail)-1]

	result = make([]interface{}, 0)
	for len(tail) > 0 {
		var ok bool
		var res interface{}
		//Check Integer
		ok, res, tail, err = ParseIfInt(tail)
		if ok && err != nil {
			return result, err
		}

		if ok {
			result = append(result, res)
			continue
		}

		//Check simple string
		ok, res, tail, err = ParseIfString(tail)
		if ok && err != nil {
			return result, err
		}

		if ok {
			result = append(result, res)
			continue
		}

		//Check simple string
		ok, res, tail, err = ParseIfError(tail)
		if ok && err != nil {
			return result, err
		}

		if ok {
			result = append(result, res)
			continue
		}

		//Check bulk string
		ok, res, tail, err = ParseIfBulkString(tail)
		if ok && err != nil {
			return result, err
		}

		if ok {
			result = append(result, res)
			continue
		}

		var arLen int64
		ok, tail, arLen, err = isArray(tail)
		if ok && err != nil {
			return result, err
		}

		if ok {
			res, tail, err = ParseArray(tail, arLen)
			if err != nil {
				return result, err
			}

			result = append(result, res)
			continue
		}
		//Must be ureacheble
		return result, errors.New("RESP parsing error. Uncknown RESP type")
	}
	return result, nil
}

func FormatInt(val int) string {
	return fmt.Sprintf("%s%d\r\n", prf_int, val)
}

func FormatBulkString(val string) string {
	return string(prf_bulk) + strconv.FormatInt(int64(utf8.RuneCountInString(val)), 10) + "\r\n" + val + "\r\n"
}

func FormatError(val error) string {
	return fmt.Sprintf("%r%s\r\n", prf_err, val)
}

func FormatString(val string) string {
	return fmt.Sprintf("%r%s\r\n", prf_str, val)
}

func FormatArray(arr []interface{}) (string, error) {
	str := fmt.Sprintf("%s%d\r\n", prf_arr, len(arr))

	for i, item := range arr {
		switch item.(type) {
		case string:
			str += FormatString(item.(string))
		case []interface{}:
			f, err := FormatArray(item.([]interface{}))
			if err != nil {
				return "", err
			}
			str += f
			break
		case error:
			str += FormatError(item.(error))
		case int:
			str += FormatInt(item.(int))
		default:
			return "", fmt.Errorf("item %d of array %q has wrong value", i, arr)
		}
	}

	return str, nil
}
