package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	_ "github.com/mattn/go-sqlite3"
)

func spawnSSH(user, ip, port string) {
	cmd := exec.Command("ssh", "-p", port, user+"@"+ip)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func createDB() {
	db, err := sql.Open("sqlite3", "remotego.db")

	if err != nil {
		panic(err)
	}

	query, err := db.Prepare("CREATE TABLE devices (device text, user text, ip text, port int)")

	if err != nil {
		panic(err)
	}

	query.Exec()
}

func updateDB(srvInfo []string) {

	if _, err := os.Stat("remotego.db"); os.IsNotExist(err) {
		fmt.Println("DB does not exist, creating now")
		createDB()
	}

	db, err := sql.Open("sqlite3", "remotego.db")

	if err != nil {
		panic(err)
	}

	query, err := db.Prepare("INSERT INTO devices VALUES (?, ?, ?, ?)")

	if err != nil {
		panic(err)
	}

	query.Exec(srvInfo[0], srvInfo[1], srvInfo[2], srvInfo[3])
}

func readDB(devicename string) (device, user, ip, port string) {
	db, _ := sql.Open("sqlite3", "remotego.db")

	if devicename == "" {
		rows, err := db.Query("SELECT * FROM devices ORDER BY device COLLATE NOCASE ASC")

		if err != nil {
			panic(err)
		}

		for rows.Next() {
			fmt.Println("===============")
			rows.Scan(&device, &user, &ip, &port)
			fmt.Printf("device : %s\n", device)
			fmt.Printf("user   : %s\n", user)
			fmt.Printf("ip     : %s\n", ip)
			fmt.Printf("port   : %s\n", port)
		}
		fmt.Println("===============")
		defer rows.Close()
	} else {
		query := "SELECT * FROM devices WHERE device='" + devicename + "'"
		rows, err := db.Query(query)

		if err != nil {
			panic(err)
		}

		defer rows.Close()

		for rows.Next() {
			rows.Scan(&device, &user, &ip, &port)
		}

		if device == "" {
			fmt.Println("[e] device not found")
			os.Exit(1)
		}
	}

	return
}

func deleteServer(device string) {
	db, err := sql.Open("sqlite3", "remotego.db")

	if err != nil {
		panic(err)
	}

	db_query := "DELETE FROM devices WHERE device='" + device + "'"
	query, err := db.Prepare(db_query)

	if err != nil {
		panic(err)
	}

	query.Exec()
}

func updateDevice(device, comp, value string) {
	db, err := sql.Open("sqlite3", "remotego.db")

	if err != nil {
		panic(err)
	}

	db_query := "UPDATE devices SET " + comp + "='" + value + "' WHERE device='" + device + "'"
	fmt.Println(db_query)
	query, err := db.Prepare(db_query)

	if err != nil {
		panic(err)
	}

	query.Exec()
}

func usage(name string) {
	fmt.Printf("usage : %s [args]\n", name)
	fmt.Println(" -a, --add <device|user|ip|port>\tadd device to list")
	fmt.Println(" -u, --update <device|component|value>\tupdate device settings")
	fmt.Println(" -c, --connect [device]\t\t\tconnect to device")
	fmt.Println(" -d, --delete [device]\t\t\tdelete device from list")
	fmt.Println(" -l, --list\t\t\t\tlist devices")
}

func main() {
	argv := os.Args
	argc := len(os.Args)

	if argc < 2 {
		usage(argv[0])
	}

	for i := 0; i < argc; i++ {
		if argv[i] == "-a" || argv[i] == "--add" {
			srvInfo := make([]string, 4)
			index := 0
			for j := 2; j <= 5; j++ {
				srvInfo[index] = argv[j]
				index++
			}
			updateDB(srvInfo)
		} else if argv[i] == "-u" || argv[i] == "--update" {
			if argc != 5 {
				usage(argv[0])
			} else {
				updateDevice(argv[2], argv[3], argv[4])
			}
		} else if argv[i] == "-c" || argv[i] == "--connect" {
			_, user, ip, port := readDB(argv[2])
			spawnSSH(user, ip, port)
		} else if argv[i] == "-d" || argv[i] == "--delete" {
			deleteServer(argv[i + 1])
		} else if argv[i] == "-l" || argv[i] == "--list" {
			if argc == 3 {
				readDB(argv[2])
			} else {
				readDB("")
			}
		} else {
			continue
		}
	}
}
