package main;

import "encoding/json"
import "fmt"
import "os"
import "sync"
import "sort"
import "io/ioutil"
import "errors"
import "time"



// Represents an entry: a timeslot that may be reserved for a user
type Entry struct{
    Year int
    Month int
    Day int
    Slot int
}

func (r *Entry) String() string{
    return fmt.Sprintf("%d.%d.%d(%d)", r.Day, r.Month, r.Year, r.Slot)
}

func (r *Entry) Equals(other Entry) bool{
    return r.Year==other.Year && r.Month==other.Month && r.Day==other.Day && r.Slot==other.Slot
}

// Represents a user and all his entries (entries)
type User struct{
    Name string
    Password string
    Entries []Entry
}

// Represents all users
type Users struct{
    users []User
    entry_to_user map[Entry]string
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

    entry_to_user:=make(map[Entry]string)
    for _,user:=range users{
        for _, entry:=range user.Entries{
            entry_to_user[entry]=user.Name
        }
    }

    return Users{users, entry_to_user, &sync.RWMutex{}}, nil
}

func new_users() Users{
    return Users{[]User{}, make(map[Entry]string), &sync.RWMutex{}}
}

func (u *Users) to_file(filename string) error{
    u.lock.Lock() // Full lock due to file access
    defer u.lock.Unlock()

    // b, err := json.Marshal(u.users)
    b, err := json.MarshalIndent(u.users, "", "    ")
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

func (u *Users) remove_old_entries(){
    year, month, day:=time.Now().Date()
    u.lock.Lock()
    defer u.lock.Unlock()

    for i:=0; i<len(u.users); i++{
        entries:=[]Entry{}
        for _,entry:=range u.users[i].Entries{
            if (entry.Year>year) ||
               (entry.Year==year && entry.Month>int(month)) ||
               (entry.Year==year && entry.Month==int(month) && entry.Day>=day) {
                entries=append(entries, entry)
            } else{
                delete(u.entry_to_user, entry)
            }
        }
        u.users[i].Entries=entries
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

    u.users=append(u.users, User{name, password, []Entry{}})

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
            for _,entry:=range user.Entries{
                delete(u.entry_to_user, entry)
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

func (u *Users) add_entry(name string, entry Entry) error{
    if entry.Year<2017 || entry.Month>12 || entry.Month<1 || entry.Day>31 || entry.Day<1 || entry.Slot>23 || entry.Slot<0{
        return errors.New("Will not add invalid entry")
    }
    u.lock.Lock()
    defer u.lock.Unlock()
    if _,ok:=u.entry_to_user[entry]; ok{
        return errors.New("Entry already exists")
    }

    added:=false
    for i:=0; i<len(u.users); i++{
        if u.users[i].Name==name{
            u.users[i].Entries=append(u.users[i].Entries, entry)
            u.entry_to_user[entry]=name
            added=true
            break
        }
    }

    if !added{
        return errors.New("User does not exist")
    }

    return nil
}

func (u *Users) remove_entry(name string, entry Entry) error{
    u.lock.Lock()
    defer u.lock.Unlock()

    removed:=false
    for i:=0; i<len(u.users); i++{
        if u.users[i].Name==name{
            entries:=[]Entry{}
            for _, entry_o:=range u.users[i].Entries{
                if !entry_o.Equals(entry){
                    entries=append(entries, entry_o)
                } else{
                    delete(u.entry_to_user, entry)
                    removed=true
                }
            }
            u.users[i].Entries=entries
            break
        }
    }

    if !removed{
        return errors.New("Could not find entry")
    }

    return nil
}

func (u *Users) remove_all_entries(){
    u.lock.Lock()
    defer u.lock.Unlock()

    for i:=0; i<len(u.users); i++{
        u.users[i].Entries=[]Entry{}
    }
}

func (u *Users) get_entries_on_day(entry_day Entry) [24]string{
    u.lock.RLock()
    defer u.lock.RUnlock()

    ret:=[24]string{}
    for i:=0; i<24; i++{
        entry_day.Slot=i
        ret[i]=u.entry_to_user[entry_day]
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