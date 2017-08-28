package main;

import "log"
import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "io/ioutil"
// import "crypto/sha256"
// import "crypto/sha512"
// import "crypto/md5"
import "strings"
import "errors"
// import "strconv"
import "sync"
// import "crypto/rand"
import "os"
import "bytes"

const(
    user_file = "data/users.json"
    root_address = "/"
    root_file = "frontend/index.html"
    get_entries_address = "/get_entries"
)

type Reservation struct{
    Year int
    Month int
    Day int
    Slot int
}

// func (r *Reservation) String() string{
//     return fmt.Sprintf("%d.%d.%d(%d)", r.Day, r.Month, r.Year, r.Slot)
// }

func (r *Reservation) Equals(other Reservation) bool{
    return r.Year==other.Year && r.Month==other.Month && r.Day==other.Day && r.Slot==other.Slot
}

type User struct{
    Name string
    Password string
    Reservations []Reservation
}

type Users struct{
    users []User
    reservation_to_user map[Reservation]string
    lock *sync.RWMutex
}

func from_file(filename string) (Users,error){
    var users []User
    content, err:=ioutil.ReadFile(filename)
    if err!=nil{
        return Users{nil,nil,nil}, err
    }

    err=json.Unmarshal(content, &users)
    if err!=nil{
        return Users{nil,nil,nil}, err
    }

    reservation_to_user:=make(map[Reservation]string)
    for _,user:=range users{
        for _, reservation:=range user.Reservations{
            reservation_to_user[reservation]=user.Name
        }
    }

    return Users{users, reservation_to_user, &sync.RWMutex{}}, nil
}

func (u *Users) to_file(filename string) error{
    u.lock.Lock() // Full lock due to file access
    defer u.lock.Unlock()

    b, err := json.Marshal(u.users)
    if err!=nil{
        return err
    }

    f, err:=os.Create(filename)
    if err!=nil{
        return err
    }
    f.Write(b)
    f.Close()
    return nil
}

func (u *Users) remove_old_reservations(){
    year, month, day:=time.Now().Date()
    u.lock.Lock()
    defer u.lock.Unlock()

    for i:=0; i<len(u.users); i++{
        reservations:=[]Reservation{}
        for _,reservation:=range u.users[i].Reservations{
            if (reservation.Year>year) ||
               (reservation.Year==year && reservation.Month>int(month)) ||
               (reservation.Year==year && reservation.Month==int(month) && reservation.Day>=day) {
                reservations=append(reservations, reservation)
            } else{
                delete(u.reservation_to_user, reservation)
            }
        }
        u.users[i].Reservations=reservations
    }
}

func (u *Users) add_user(name, password string) error{
    u.lock.Lock()
    defer u.lock.Unlock()

    for _,user:=range u.users{
        if user.Name==name{
            return errors.New("A user with that name already exists")
        }
    }

    u.users=append(u.users, User{name, password, []Reservation{}})

    return nil
}

func (u *Users) remove_user(name string) error{
    u.lock.Lock()
    defer u.lock.Unlock()

    users:=[]User{}
    removed:=false
    for _,user:=range u.users{
        if user.Name!=name{
            users=append(users, user)
        } else{
            for _,reservation:=range user.Reservations{
                delete(u.reservation_to_user, reservation)
            }
            removed=true
        }
    }

    if !removed{
        return errors.New("User does not exist")
    }

    u.users=users
    return nil
}

func (u *Users) add_reservation(name string, reservation Reservation) error{
    if reservation.Year<2017 || reservation.Month>12 || reservation.Month<1 || reservation.Day>31 || reservation.Day<1 || reservation.Slot>23 || reservation.Slot<0{
        return errors.New("Will not add invalid reservation")
    }
    u.lock.Lock()
    defer u.lock.Unlock()
    if _,ok:=u.reservation_to_user[reservation]; ok{
        return errors.New("Reservation already exists")
    }

    added:=false
    for i:=0; i<len(u.users); i++{
        if u.users[i].Name==name{
            u.users[i].Reservations=append(u.users[i].Reservations, reservation)
            u.reservation_to_user[reservation]=name
            added=true
            break
        }
    }

    if !added{
        return errors.New("User does not exist")
    }

    return nil
}

func (u *Users) remove_reservation(name string, reservation Reservation) error{
    u.lock.Lock()
    defer u.lock.Unlock()

    removed:=false
    for i:=0; i<len(u.users); i++{
        if u.users[i].Name==name{
            reservations:=[]Reservation{}
            for _, reservation_o:=range u.users[i].Reservations{
                if !reservation_o.Equals(reservation){
                    reservations=append(reservations, reservation_o)
                } else{
                    delete(u.reservation_to_user, reservation)
                    removed=true
                }
            }
            u.users[i].Reservations=reservations
            break
        }
    }

    if !removed{
        return errors.New("Could not find reservation")
    }

    return nil
}

func (u *Users) get_reservations_on_day(reservation_day Reservation) [24]string{
    u.lock.RLock()
    defer u.lock.RUnlock()

    ret:=[24]string{}
    for i:=0; i<24; i++{
        reservation_day.Slot=i
        ret[i]=u.reservation_to_user[reservation_day]
    }

    return ret
}


// func fun(){
//     users:=[]User{
//     User{"502", "toni", []Reservation{Reservation{2017, 9, 28, 4}}},
//     User{"504", "sabrina", []Reservation{Reservation{2017, 9, 28, 4}, Reservation{2017, 9, 28, 5}, Reservation{2017, 9, 28, 6}}},
//     }

    // b, err := json.Marshal(users)
    // if err!=nil{
    //     panic(err.Error())
    // }

    // f, err:=os.Create("hello.json")
    // if err!=nil{
    //     panic(err.Error())
    // }
    // f.Write(b)
    // f.Close()
// }


// TODO: always iterate over index range when changing stuff
// TODO: Look at every function and make sure the map is updated correctly
// TODO: syncrhonize it all (channels maybe?) (send pointer for return values?)
// TODO: when a user is removed, all his reservations should be removed from map
// strings.ToLower("Gopher")
// delete(m, "route")
func main() {
    users,_:=from_file("asd.json")

    fmt.Println(users)
    // // users.remove_old_reservations()
    // // fmt.Println(users)
    // // fmt.Println(users.get_reservation_string_user_name_map())
    // // users.add_user("510", "felix")
    // // users.remove_user("510")
    // users.add_reservation("502", Reservation{2017, 8, 28, 10}, users.get_reservation_string_user_name_map())
    // fmt.Println(users)
    // users.to_file("asd.json")
    // return

    // b, err := json.Marshal(logins)
    // if err!=nil{
    //     panic(err.Error())
    // }

    // f, err:=os.Create("hello.json")
    // if err!=nil{
    //     panic(err.Error())
    // }
    // f.Write(b)
    // f.Close()


























    bootstrap_files:=[]string{
        "bootstrap/css/bootstrap.min.css",
        "bootstrap/fonts/glyphicons-halflings-regular.ttf",
        "bootstrap/fonts/glyphicons-halflings-regular.woff",
        "bootstrap/fonts/glyphicons-halflings-regular.woff2",
    }

    mux:=http.NewServeMux()
    for _,bootstrap_file:=range bootstrap_files{
        bootstrap_file_address_bytes:=make([]byte, len(bootstrap_file)+1)
        copy(bootstrap_file_address_bytes[1:], []byte(bootstrap_file))
        bootstrap_file_address_bytes[0]=([]byte("/"))[0]
        bootstrap_file_address:=string(bootstrap_file_address_bytes)

        mimetype:=func() string{
            if strings.HasSuffix(bootstrap_file, ".css"){
                return "text/css"
            } else if strings.HasSuffix(bootstrap_file, ".js") {
                return "application/javascript"
            } else{
                return "text/plain"
            }
        }()

        bootstrap_file_copy:=bootstrap_file
        mux.HandleFunc(bootstrap_file_address, func (w http.ResponseWriter, r *http.Request){
            w.Header().Set("Content-Type", mimetype)
            http.ServeFile(w, r, bootstrap_file_copy)
        })
    }

    mux.HandleFunc(root_address, func (w http.ResponseWriter, r *http.Request){
        http.ServeFile(w, r, root_file)
    })

    mux.HandleFunc(get_entries_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="POST"{
            days_in_the_future:=func() int{
                if r.Body==nil{
                    return 0
                }
                buf:=new(bytes.Buffer)
                buf.ReadFrom(r.Body)
                r.Body.Close()

                var Days_in_the_future struct{
                    Days_in_the_future int
                }

                // var Days_in_the_future days_in_the_future
                err:=json.Unmarshal(buf.Bytes(), &Days_in_the_future)
                if err!=nil{
                    return 0
                }

                return Days_in_the_future.Days_in_the_future
            }()

            fmt.Println("days_in_the_future: ", days_in_the_future)

            var to_send struct{
                Date string
                Entries [24] string
            }

            now:=time.Now().AddDate(0,0,days_in_the_future)
            year, month, day:=now.Date()
            weekday:=now.Weekday()
            // to_send.Date=fmt.Sprintf("%s, %d of %s of %d", weekday.String(), day, month.String(), year)
            to_send.Date=fmt.Sprintf("%s, %d of %s", weekday.String(), day, month.String())
            entries:=users.get_reservations_on_day(Reservation{year, int(month), day, 0})
            copy(to_send.Entries[:], entries[:])
            // fmt.Println(to_send)

            json_entries_data,err:=json.Marshal(to_send)
            if err!=nil{
                log.Println(err)
                http.Error(w, "Internal Error", http.StatusInternalServerError)
                return
            }

            w.Write(json_entries_data)

        } else {
            http.Error(w, "Request must be POST.", http.StatusBadRequest)
            return
        }
    })

    log.Println("Start server")
    if err:=http.ListenAndServe(":8000", mux);err!=nil{
        log.Println(err)
    }
}