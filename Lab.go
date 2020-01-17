package main

import (
	"net"
	"fmt"
	"bufio"
	"strings"
	"time"
	"strconv"
	"math/rand"
	"flag"
)

func SessionKey() string { //генерация ключа
	res := ""
	for i := 0; i < 10; i++{ res += string(strconv.Itoa(int(9 * rand.Float64()) + 1)[0]) }
	return res
}

func HashString() string {
	hres := ""
	for i := 0; i < 5 ; i++{ hres += strconv.Itoa(int(int((6 * rand.Float64()) + 1))) }
	return hres
}

type _protector struct { //защита от неавторизированного пользователя
	__hash string
}

func (self _protector) HashCalculation (key string, val int) string { //вычисление хеша/хэша
	if val == 1 {
		res := ""
		ret := ""
		for idx := 0; idx < 5; idx++ {
			res += string(key[idx])        }
		i := strconv.Atoi(res)
		res = "00" + strconv.Itoa(i % 97)
		for idx := len(res) - 2; idx < len(res); idx++ {
			ret += string(res[idx])
		}
		return ret
	}
	if val == 2 {
		res := ""
		for idx := 0; idx < len(key); idx++{
			result += string(key[len(key) - idx - 1])
		}
		return res
	}
	if val == 3 {
		res := ""
		ret := ""
		for idx := 0; idx < 5; idx++ {
			res += string(key[idx])
		}
		for idx := 5; idx < len(key); idx++ {
			ret += string(key[idx])
		}
		return ret + res
	}
	if val == 4 {
		res:= 0
		for idx := 1; idx < 8; idx++ {
			num :=  strconv.Atoi(string(key[idx]))
			res += num + 41
		}
		return strconv.Itoa(res)
	}
	if val == 5 {
		var ch string
		res := 0
		for idx := 0; idx < len(key); idx++ {
			ch = string(int(int(key[idx]) ^ 43))
			if err := strconv.Atoi(ch); err != nil {
				ch = string(int(ch[0]))
			}
			num := strconv.Atoi(ch)
			res += num
		}
		return strconv.Itoa(res)
	}
	res:= strconv.Atoi(key)
	return strconv.Itoa(res + val)
}

func (self _protector) NextKey(session_key string) string { //генерация следующего ключа
	if self.__hash == "" {
		fmt.Println("hash is empty")
		return SessionKey()
	}
	for idx := 0; idx < len(self.__hash); idx++ {
		i := string(self.__hash[idx])
		if _, err := strconv.Atoi(i); err != nil {
			fmt.Println("Here is letter")
			return SessionKey()
		}
	}
	result := 0
	ret := ""
	for idx := 0; idx < len(self.__hash); idx++ {
		num, _ := strconv.Atoi(string(self.__hash[idx]))
		k, _ := strconv.Atoi(self.HashCalculation(session_key, num))
		result += k
	}
	for idx := 0; idx < 10 && idx < len(strconv.Itoa(result)); idx++ {
		ret += string((strconv.Itoa(result))[idx])
	}
	m := ""
	ret = "0000000000" + ret
	for idx := len(ret) - 10; idx < len(ret); idx++ {
		m += string(ret[idx])
	}
	return m
}

func run_connection(conn *net.Conn, id int, point *int) {

	// run loop forever (or until ctrl-c)
	text, serr := bufio.NewReader(*conn).ReadString('\n')
	if serr == nil {
		serv_hash_string := ""
		key1 := ""
		for i := 0; i < 5; i++ {
			serv_hash_string += string(text[i])
		}
		for i := 5; i < 15; i++ {
			key1 += string(text[i])
		}
		fmt.Println(serv_hash_string, key1)
		server_protector := _protector{strings.Replace(serv_hash_string, "\n", "", -1)}
		key2 := server_protector.NextKey(key1)
		(*conn).Write([]byte(key2 + "\n"))
		for {
			// will listen for message to process ending in newline (\n)
			message, err := bufio.NewReader(*conn).ReadString('\n')
			if err == nil {
				key1 = ""
				text = ""
				for i := len(message) - 11; i < len(message) - 1; i++ {
					key1 += string(message[i])
				}
				for i := 0; i < len(message) - 11; i++ {
					text += string(message[i])
				}
				// output message received
				fmt.Println("Client ( id = ", id, " ) message: ", string(text), "key: ", key1)
				// sample process for string received
				newmessage := strings.ToUpper(text)
				key2 = server_protector.NextKey(strings.Replace(key1, "\n", "", -1))
				fmt.Print("New key: ", key2, "\n")
				// send new string back to client
				(*conn).Write([]byte(newmessage + key2 + "\n"))
			}else{
				(*conn).Close()
				*point -= 1
				fmt.Println("Client ( id =", id, ") Disconnected!")
				break
			}
		}
	}else{
		(*conn).Close()
		*point -= 1
		fmt.Println("Client ( id =", id, ") Disconnected!")
	}
}

func main() {
	port := flag.String("port", ":8081", "a server listening port")
	IP := flag.String("ip:port", "", "a client connection port")
	n := flag.Int("n", 100, "a number of simultaneous connections")
	flag.Parse()
	if *IP == "" {
		var id = 1
		ln := net.Listen("tcp", *port)
		point := 1
		for {
			conn := ln.Accept()
			if point <= *n {
				point += 1
				fmt.Println("New client ( id =", id, ") Connected!")
				go run_connection(&conn, id, &point)
				id += 1
			}else{conn.Close()}
		}
	}else{
		rand.Seed(time.Now().UnixNano())
		conn, err := net.Dial("tcp", *IP)
		if err != nil {
			fmt.Println("Server not found. ")
		}else{
			cl_hash_string := HashString()
			key1 := SessionKey()
			fmt.Print(cl_hash_string + "\n")
			fmt.Fprintf(conn, cl_hash_string + key1 + "\n")
			client_protector := _protector{cl_hash_string}
			key2, err := bufio.NewReader(conn).ReadString('\n')
			if err != nil {
				fmt.Println("Server closed connection. ")
			}
			key1 = client_protector.NextKey(key1)
			key1 = client_protector.NextKey(key1)
			for {
				fmt.Print("Text to send: ")
				text := ""

				fmt.Fprintf(conn, strings.Replace(text, "\n", "", -1) + key1 + "\n")

				fmt.Println("Waiting for answer...")
				message, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Println("Server closed connection. Try later again.")
				}
				key2 = ""
				text = ""
				for i := len(message) - 11; i < len(message) - 1; i++ {
					key2 += string(message[i])
				}
				for i := 0; i < len(message) - 11; i++ {
					text += string(message[i])
				}
				key1 = client_portector.NextKey(key1)
				fmt.Println("Message from server: " + text, "key: ", key1, " ", key2)
				key1 = client_portector.NextKey(key1)
			}
		}
	}
}