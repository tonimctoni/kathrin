package main;

import "testing"
import "io/ioutil"
import "os"
import "time"

func TestEntry(t *testing.T){
    entry:=Entry{1,2,3,4}
    if entry.String()!="3.2.1(4)"{
        t.Error()
    }

    if !entry.Equals(entry){
        t.Error()
    }

    if !entry.Equals(Entry{1,2,3,4}){
        t.Error()
    }

    other_entries:=[]Entry{
        Entry{4,2,3,4},
        Entry{1,3,3,4},
        Entry{1,2,2,4},
        Entry{1,2,3,1},
    }

    for _,other_entry := range other_entries{
        if entry.Equals(other_entry) || other_entry.Equals(entry){
            t.Error()
        }
    }
}

func TestUsersLen(t *testing.T){
    users:=new_users()
    if users.Len()!=0{
        t.Error()
    }

    users.users=append(users.users, User{"", "", []Entry{}})
    if users.Len()!=1{
        t.Error()
    }
}

func TestUsersLess(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"a", "", []Entry{}})
    users.users=append(users.users, User{"b", "", []Entry{}})
    if users.Less(0,1)!=true{
        t.Error()
    }

    if users.Less(1,0)!=false{
        t.Error()
    }
}

func TestUsersSwap(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"a", "", []Entry{}})
    users.users=append(users.users, User{"b", "", []Entry{}})
    users.Swap(0,1)

    if users.users[0].Name!="b" || users.users[1].Name!="a"{
        t.Error()
    }
}

func TestUsersSort(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"b", "", []Entry{}})
    users.users=append(users.users, User{"a", "", []Entry{}})
    users.Sort()

    if users.users[0].Name!="a" || users.users[1].Name!="b"{
        t.Error()
    }
}

func TestUsersFrom_file(t *testing.T){
    json:=`
    [
        {
            "Name": "a",
            "Password": "ap",
            "Entries": [
                {
                    "Year": 1,
                    "Month": 2,
                    "Day": 3,
                    "Slot": 4
                }
            ]
        },
        {
            "Name": "b",
            "Password": "bp",
            "Entries": []
        }
    ]
    `

    err:=ioutil.WriteFile("DELETEME.json", []byte(json), 0644)
    if err!=nil{
        panic("Could not create temporary file")
    }

    users, err:=from_file("DELETEME.json")
    if err!=nil{
        t.Error(err)
        return
    }

    if users.Len()!=2{
        t.Error(err)
    }

    if users.users[0].Name!="a"{
        t.Error()
    }

    if users.users[0].Password!="ap"{
        t.Error()
    }

    if !users.users[0].Entries[0].Equals(Entry{1,2,3,4}){
        t.Error()
    }

    err=os.Remove("DELETEME.json")
    if err!=nil{
        panic("Could not remove temporary file")
    }
}

func TestUsersTo_file(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"a", "ap", []Entry{}})
    users.users=append(users.users, User{"b", "bp", []Entry{Entry{1,2,3,4}, Entry{5,6,7,8}}})

    users.to_file("DELETEME.json")
    users, err:=from_file("DELETEME.json")
    if err!=nil{
        t.Error(err)
        return
    }

    if users.Len()!=2{
        t.Error()
        return
    }

    if users.users[0].Name!="a" || users.users[0].Password!="ap"{
        t.Error()
    }

    if len(users.users[1].Entries)!=2{
        t.Error()
        return
    }

    if !users.users[1].Entries[1].Equals(Entry{5,6,7,8}){
        t.Error()
    }

    if len(users.entry_to_user)!=2{
        t.Error()
        return
    }

    if users.entry_to_user[Entry{1,2,3,4}]!="b"{
        t.Error()
    }

    err=os.Remove("DELETEME.json")
    if err!=nil{
        panic("Could not remove temporary file")
    }
}

func TestUsersAs_json(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"a", "ap", []Entry{}})
    users.users=append(users.users, User{"b", "bp", []Entry{Entry{1,2,3,4}, Entry{5,6,7,8}}})

    json,err:=users.as_json()
    if err!=nil{
        t.Error(err)
        return
    }

    err=ioutil.WriteFile("DELETEME.json", []byte(json), 0644)
    if err!=nil{
        panic("Could not create temporary file")
    }

    users, err=from_file("DELETEME.json")
    if err!=nil{
        t.Error(err)
        return
    }

    if users.Len()!=2{
        t.Error()
        return
    }

    if users.users[0].Name!="a" || users.users[0].Password!="ap"{
        t.Error()
    }

    if len(users.users[1].Entries)!=2{
        t.Error()
        return
    }

    if !users.users[1].Entries[1].Equals(Entry{5,6,7,8}){
        t.Error()
    }

    if len(users.entry_to_user)!=2{
        t.Error()
        return
    }

    if users.entry_to_user[Entry{1,2,3,4}]!="b"{
        t.Error()
    }

    err=os.Remove("DELETEME.json")
    if err!=nil{
        panic("Could not remove temporary file")
    }
}

func TestUsersRemove_old_entries(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"b", "bp", []Entry{}})

    year, month, day:=time.Now().Date()
    users.users[0].Entries=append(users.users[0].Entries, Entry{int(year), int(month), int(day)+1, 2})
    users.users[0].Entries=append(users.users[0].Entries, Entry{int(year), int(month), int(day)-1, 2})
    users.users[0].Entries=append(users.users[0].Entries, Entry{int(year), int(month)+1, int(day), 2})
    users.users[0].Entries=append(users.users[0].Entries, Entry{int(year), int(month)-1, int(day), 2})
    users.users[0].Entries=append(users.users[0].Entries, Entry{int(year)+1, int(month), int(day), 2})
    users.users[0].Entries=append(users.users[0].Entries, Entry{int(year)-1, int(month), int(day), 2})
    users.remove_old_entries()

    if len(users.users[0].Entries)!=3{
        t.Error()
    }
}

func TestUsersAdd_user(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"b", "bp", []Entry{}})
    users.add_user("name", "password")

    if len(users.users)!=2{
        t.Error()
        return
    }

    if users.users[1].Name!="name" || users.users[1].Password!="password"{
        t.Error()
    }
}

func TestUsersRemove_user(t *testing.T){
    users:=new_users()
    users.users=append(users.users, User{"b", "bp", []Entry{}})
    users.add_user("name", "password")

    err:=users.remove_user("name")
    if err!=nil{
        t.Error(err)
        return
    }

    if users.Len()!=1{
        t.Error(err)
    }

    err=users.remove_user("name")
    if err==nil{
        t.Error()
        return
    }
}