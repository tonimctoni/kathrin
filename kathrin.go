package main;

import "log"
import "encoding/json"
import "net/http"
import "time"
import "fmt"
import "io/ioutil"
import "strings"
import "errors"
import "sync"
import "os"
import "sort"

const(
    admin_user_name = "admin"
    user_file = "data/users.json"
    logfile_filename = "logfile.log"

    root_address = "/"
    root_file = "frontend/index.html"

    change_password_address = "/change_password"
    change_password_file = "frontend/change_password.html"

    see_all_address = "/see_all"
    remove_old_addresses = "/remove_old"
    admin_file = "frontend/admin.html"

    get_entries_address = "/get_entries"
    add_entry_address = "/add_entry"
    remove_entry_address = "/remove_entry"

    character_whitelist = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    max_password_length = 32
    min_password_length = 4
)

// Represents an entry: a timeslot that may be reserved for a user
type Reservation struct{
    Year int
    Month int
    Day int
    Slot int
}

func (r *Reservation) String() string{
    return fmt.Sprintf("%d.%d.%d(%d)", r.Day, r.Month, r.Year, r.Slot)
}

func (r *Reservation) Equals(other Reservation) bool{
    return r.Year==other.Year && r.Month==other.Month && r.Day==other.Day && r.Slot==other.Slot
}

// Represents a user and all his entries (reservations)
type User struct{
    Name string
    Password string
    Reservations []Reservation
}

// Represents all users
type Users struct{
    users []User
    reservation_to_user map[Reservation]string
    lock *sync.RWMutex
}

func (u Users) Len() int{
    return len(u.users)
}

func (u Users) Less(i, j int) bool{
    return u.users[i].Name<u.users[j].Name
}

func (u Users) Swap(i, j int){
    aux:=u.users[i]
    u.users[i]=u.users[j]
    u.users[j]=aux
}

func (u Users) Sort(){
    u.lock.Lock()
    defer u.lock.Unlock()
    sort.Sort(u)
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

func new_users() Users{
    return Users{[]User{}, make(map[Reservation]string), &sync.RWMutex{}}
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

func (u *Users) as_json() ([]byte, error){
    u.lock.RLock()
    defer u.lock.RUnlock()

    b, err := json.MarshalIndent(u.users, "", "    ")
    if err!=nil{
        return []byte{}, err
    }

    return b, nil
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

func (u *Users) remove_all_reservations(){
    u.lock.Lock()
    defer u.lock.Unlock()

    for i:=0; i<len(u.users); i++{
        u.users[i].Reservations=[]Reservation{}
    }
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

func (u *Users) get_users_password(user string) (string, error){
    u.lock.RLock()
    defer u.lock.RUnlock()

    for _,_user:=range u.users{
        if _user.Name==user{
            return _user.Password, nil
        }
    }

    return "", errors.New("User with that name does not exist")
}

func (u *Users) change_password(user, password, new_password string) error{
    u.lock.Lock()
    defer u.lock.Unlock()

    for i:=0; i<len(u.users); i++{
        if u.users[i].Name==user{
            if u.users[i].Password!=password{
                return errors.New("Incorrect password")
            }

            u.users[i].Password=new_password
            return nil
        }
    }

    return errors.New("User does not exist")
}

// TODO: add https
func main() {
    // users:=new_users()
    // rooms:=[]string{
    //     "101","102","103","104","105","106","107","108","109","110",
    //     "201","202","203","204","205","206","207","208","209","210",
    //     "301","302","303","304","305","306","307","308","309","310",
    //     "401","402","403","404","405","406","407","408","409","410",
    //     "501","502","503","504","505","506","507","508","509","510",
    // }

    // for _,room:=range rooms{
    //     users.add_user(room, "password")
    // }

    // fmt.Println(users.users)
    // users.to_file(user_file)
    // return

    logfile_f, err := os.OpenFile(logfile_filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer logfile_f.Close()
    log.SetOutput(logfile_f)

    users, err:=from_file(user_file)
    if err!=nil{
        log.Fatal(err)
    }
    users.add_user(admin_user_name, "password")
    users.Sort()

    bootstrap_files:=[]string{
        "bootstrap/css/bootstrap.min.css",
        "bootstrap/js/bootstrap.min.js",
        "bootstrap/js/jquery.min.js",
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

    // Sends all enties (reservations) of a certain day to the frontend
    mux.HandleFunc(get_entries_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="POST"{
            days_in_the_future:=func() int{
                if r.Body==nil{
                    return 0
                }

                content, err:=ioutil.ReadAll(r.Body)
                r.Body.Close()
                if r.Body==nil{
                    return 0
                }

                var Days_in_the_future struct{
                    Days_in_the_future int
                }

                err=json.Unmarshal(content, &Days_in_the_future)
                if err!=nil{
                    return 0
                }

                return Days_in_the_future.Days_in_the_future
            }()

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

            json_entries_data,err:=json.Marshal(to_send)
            if err!=nil{
                log.Println(err)
                http.Error(w, "Internal Error", http.StatusInternalServerError)
                return
            }

            w.Header().Set("Content-Type", "application/json")
            w.Write(json_entries_data)

        } else {
            http.Error(w, "Request must be POST.", http.StatusBadRequest)
            return
        }
    })

    // Processes request to add an entry (reservation)
    mux.HandleFunc(add_entry_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="POST"{
            type AddEntryRequestData struct{
                Days_in_the_future int
                Date string
                Active_entry int
                Name string
                Password string
            }

            // Get request data
            add_entry_request_data, err:=func() (AddEntryRequestData, error){
                add_entry_request_data:=AddEntryRequestData{}
                if r.Body==nil{
                    return add_entry_request_data, errors.New("No body")
                }

                content, err:=ioutil.ReadAll(r.Body)
                r.Body.Close()
                if r.Body==nil{
                    return add_entry_request_data, errors.New("Could not read body")
                }

                err=json.Unmarshal(content, &add_entry_request_data)
                if err!=nil{
                    log.Println(err)
                    return add_entry_request_data, errors.New("Could not read request")
                }

                return add_entry_request_data, nil
            }()

            if err!=nil{
                http.NotFound(w, r)
                return
            }

            var to_send struct{
                Return_code int
            }

            // Get password, on error send error code
            password, err:=users.get_users_password(add_entry_request_data.Name)
            if err!=nil{
                to_send.Return_code=1
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // See if password is correct, on error send error code
            if add_entry_request_data.Password!=password{
                to_send.Return_code=2
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // See if dates are inconsistent
            now:=time.Now().AddDate(0,0,add_entry_request_data.Days_in_the_future)
            year, month, day:=now.Date()
            weekday:=now.Weekday()
            if fmt.Sprintf("%s, %d of %s", weekday.String(), day, month.String())!=add_entry_request_data.Date{
                to_send.Return_code=3
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // Try to add the actual reservation
            new_reservation:=Reservation{year, int(month), day, add_entry_request_data.Active_entry}
            err=users.add_reservation(add_entry_request_data.Name, new_reservation)
            if err!=nil{
                to_send.Return_code=4
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // Save changes to file
            err=users.to_file(user_file)
            if err!=nil{
                log.Println(err)
            }
            log.Println("Entry added:", add_entry_request_data.Name, new_reservation.String())

            // If the program got here, the reservation was added correctly. Send good return code
            to_send.Return_code=20
            json_to_send,err:=json.Marshal(to_send)
            if err!=nil{
                http.NotFound(w, r)
                return
            }
            w.Header().Set("Content-Type", "application/json")
            w.Write(json_to_send)
            return

        } else {
            http.Error(w, "Request must be POST.", http.StatusBadRequest)
            return
        }
    })

    // Processes request to remove an entry (reservation)
    mux.HandleFunc(remove_entry_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="POST"{
            type RemoveEntryRequestData struct{
                Days_in_the_future int
                Date string
                Active_entry int
                Name string
                Password string
            }

            // Get request data
            remove_entry_request_data, err:=func() (RemoveEntryRequestData, error){
                remove_entry_request_data:=RemoveEntryRequestData{}
                if r.Body==nil{
                    return remove_entry_request_data, errors.New("No body")
                }

                content, err:=ioutil.ReadAll(r.Body)
                r.Body.Close()
                if r.Body==nil{
                    return remove_entry_request_data, errors.New("Could not read body")
                }

                err=json.Unmarshal(content, &remove_entry_request_data)
                if err!=nil{
                    log.Println(err)
                    return remove_entry_request_data, errors.New("Could not read request")
                }

                return remove_entry_request_data, nil
            }()

            if err!=nil{
                http.NotFound(w, r)
                return
            }

            var to_send struct{
                Return_code int
            }

            // Get password, on error send error code
            password, err:=users.get_users_password(remove_entry_request_data.Name)
            if err!=nil{
                to_send.Return_code=1
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Write(json_to_send)
                return
            }

            // See if password is correct, on error send error code
            if remove_entry_request_data.Password!=password{
                to_send.Return_code=2
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // See if dates are inconsistent
            now:=time.Now().AddDate(0,0,remove_entry_request_data.Days_in_the_future)
            year, month, day:=now.Date()
            weekday:=now.Weekday()
            if fmt.Sprintf("%s, %d of %s", weekday.String(), day, month.String())!=remove_entry_request_data.Date{
                to_send.Return_code=3
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // Try to add the actual reservation
            reservation_to_remove:=Reservation{year, int(month), day, remove_entry_request_data.Active_entry}
            err=users.remove_reservation(remove_entry_request_data.Name, reservation_to_remove)
            if err!=nil{
                to_send.Return_code=4
                json_to_send,err:=json.Marshal(to_send)
                if err!=nil{
                    http.NotFound(w, r)
                    return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(json_to_send)
                return
            }

            // Save changes to file
            err=users.to_file(user_file)
            if err!=nil{
                log.Println(err)
            }
            log.Println("Entry removed:", remove_entry_request_data.Name, reservation_to_remove.String())

            // If the program got here, the reservation was removed correctly. Send good return code
            to_send.Return_code=21
            json_to_send,err:=json.Marshal(to_send)
            if err!=nil{
                http.NotFound(w, r)
                return
            }
            w.Header().Set("Content-Type", "application/json")
            w.Write(json_to_send)
            return

        } else {
            http.Error(w, "Request must be POST.", http.StatusBadRequest)
            return
        }
    })

    // Processes request to change a users password
    mux.HandleFunc(change_password_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="GET"{
            http.ServeFile(w, r, change_password_file)
            return
        } else if r.Method=="POST"{
            name:=r.FormValue("name")
            password:=r.FormValue("password")
            new_password1:=r.FormValue("new_password1")
            new_password2:=r.FormValue("new_password2")

            // If the new password and the new password re-entry are inconsistent
            if new_password1!=new_password2{
                http.Error(w, "Must entry the same password in both \"New Password\" fields", http.StatusBadRequest)
                return
            }

            // If the new password's length is not within bounds
            if len(new_password1)<min_password_length || len(new_password1)>max_password_length{
                http.Error(w, fmt.Sprintf("New password's length must be between %d and %d", min_password_length, max_password_length), http.StatusBadRequest)
                return
            }

            // Chech wheter only whitelisted characters are contained in the new password
            password_has_whitelisted_chars_only:=func() bool{
                for _,password_char:=range new_password1{
                    if !strings.ContainsAny(string(password_char), character_whitelist){
                        return false
                    }
                }
                return true
            }()
            // If there are non-whitelisted characters in the new password
            if !password_has_whitelisted_chars_only{
                http.Error(w, fmt.Sprintf("New password may only have allowed characters (%s)", character_whitelist), http.StatusBadRequest)
                return
            }

            // Do the actual password change
            err:=users.change_password(name, password, new_password1)
            if err!=nil{
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }

            // Save changes to file
            err=users.to_file(user_file)
            if err!=nil{
                log.Println(err)
            }
            log.Println("Password changed by user:", name)

            w.Write([]byte("Password changed successfully"))
            return
        } else {
            http.Error(w, "Request must be GET or POST.", http.StatusBadRequest)
            return
        }
    })

    // Processes request from admin to see all data (including all passwords)
    mux.HandleFunc(see_all_address, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="GET"{
            http.ServeFile(w, r, admin_file)
            return
        } else if r.Method=="POST"{
            admin_password:=r.FormValue("admin_password")
            admin_password_, err:=users.get_users_password(admin_user_name)
            if err!=nil{
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            // If the enetered password is not the admin's
            if admin_password_!=admin_password{
                http.Error(w, "Wrong password", http.StatusUnauthorized)
                return
            }

            // Get all data (including all passwords) in readable json format
            json_users,err:=users.as_json()
            if err!=nil{
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            w.Header().Set("Content-Type", "application/json")
            w.Write(json_users)
            return
        } else {
            http.Error(w, "Request must be GET or POST.", http.StatusBadRequest)
            return
        }
    })

    // Processes request from admin to remove old entries
    mux.HandleFunc(remove_old_addresses, func (w http.ResponseWriter, r *http.Request){
        if r.Method=="GET"{
            http.ServeFile(w, r, admin_file)
            return
        } else if r.Method=="POST"{
            admin_password:=r.FormValue("admin_password")
            admin_password_, err:=users.get_users_password(admin_user_name)
            if err!=nil{
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            if admin_password_!=admin_password{
                http.Error(w, "Wrong password", http.StatusUnauthorized)
                return
            }

            users.remove_old_reservations()
            // Save changes to file
            err=users.to_file(user_file)
            if err!=nil{
                log.Println(err)
            }
            log.Println("Old entries removed")
            w.Write([]byte("Old entries removed"))
            return
        } else {
            http.Error(w, "Request must be GET or POST.", http.StatusBadRequest)
            return
        }
    })

    log.Println("Server started")
    if err:=http.ListenAndServe(":8000", mux);err!=nil{
        log.Fatal(err)
    }
}
