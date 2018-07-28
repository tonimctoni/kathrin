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

    // Test the case that there is an error in the json file
    err=ioutil.WriteFile("DELETEME.json", []byte(json+"[error]L"), 0644)
    if err!=nil{
        panic("Could not create temporary file")
    }

    _, err=from_file("DELETEME.json")
    if err==nil{
        t.Error()
    }

    err=os.Remove("DELETEME.json")
    if err!=nil{
        panic("Could not remove temporary file")
    }

    // Test the case that there is no json file
    _, err=from_file("DELETEME.json")
    if err==nil{
        t.Error()
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
    err:=users.add_user("name", "password")
    if err!=nil{
        t.Error()
        return
    }

    if len(users.users)!=2{
        t.Error()
    }

    if users.users[1].Name!="name" || users.users[1].Password!="password"{
        t.Error()
    }

    err=users.add_user("name", "password")
    if err==nil{
        t.Error()
    }

    if len(users.users)!=2{
        t.Error()
    }

    err=users.add_user("", "password")
    if err==nil{
        t.Error()
    }

    err=users.add_user("othername", "")
    if err==nil{
        t.Error()
    }

    if len(users.users)!=2{
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


func TestUsersAdd_entry(t *testing.T){
    users:=new_users()
    users.add_user("name", "password")

    if len(users.users[0].Entries)!=0 || len(users.entry_to_user)!=0{
        t.Error()
    }

    if users.add_entry("gnome", Entry{2018, 7, 28, 2})==nil{
        t.Error()
    }

    if users.add_entry("name", Entry{2018, 7, 28, 2})!=nil{
        t.Error()
    }

    if len(users.users[0].Entries)!=1 || len(users.entry_to_user)!=1{
        t.Error()
    }

    if users.add_entry("name", Entry{2018, 7, 28, 2})==nil{
        t.Error()
    }

    // Test invalid entries
    if users.add_entry("name", Entry{2018, 7, 28, -1})==nil ||
    users.add_entry("name", Entry{2018, 7, 28, 24})==nil ||
    users.add_entry("name", Entry{2018, 7, 32, 2})==nil ||
    users.add_entry("name", Entry{2018, 7, 0, 2})==nil ||
    users.add_entry("name", Entry{2018, 0, 28, 2})==nil ||
    users.add_entry("name", Entry{2018, 13, 28, 2})==nil ||
    users.add_entry("name", Entry{2016, 7, 28, 2})==nil{
        t.Error()
    }
}

func TestUsersRemove_entry(t *testing.T){
    users:=new_users()
    users.add_user("name", "password")
    users.add_entry("name", Entry{2018, 7, 28, 2})

    if users.remove_entry("gnome", Entry{2018, 7, 28, 2})==nil{
        t.Error()
    }

    if users.remove_entry("name", Entry{2018, 7, 28, 3})==nil ||
    users.remove_entry("name", Entry{2018, 7, 27, 2})==nil ||
    users.remove_entry("name", Entry{2018, 6, 28, 2})==nil ||
    users.remove_entry("name", Entry{2019, 7, 28, 2})==nil{
        t.Error()
    }

    if users.remove_entry("name", Entry{2018, 7, 28, 2})!=nil{
        t.Error()
    }

    if len(users.users[0].Entries)!=0 || len(users.entry_to_user)!=0{
        t.Error()
    }
}

func TestUsersRemove_all_entries(t *testing.T){
    users:=new_users()
    users.add_user("name", "password")
    users.add_entry("name", Entry{2018, 7, 28, 2})

    if len(users.users[0].Entries)==0 || len(users.entry_to_user)==0{
        t.Error()
    }

    users.remove_all_entries()

    if len(users.users[0].Entries)!=0 || len(users.entry_to_user)!=0{
        t.Error()
    }
}

func TestUsersGet_entries_on_day(t *testing.T){
    users:=new_users()
    users.add_user("name", "password")
    users.add_entry("name", Entry{2018, 7, 28, 2})
    users.add_entry("name", Entry{2018, 7, 29, 3})

    for i, entry_string:=range users.get_entries_on_day(Entry{2018, 7, 28, 0}){
        if (((entry_string=="name")!=(i==2)) || ((entry_string=="") != (i!=2))){
            t.Error()
        }
    }

    for _, entry_string:=range users.get_entries_on_day(Entry{2018, 7, 30, 0}){
        if entry_string!=""{
            t.Error()
        }
    }

}

func TestUsersGet_users_password(t *testing.T){
    users:=new_users()
    users.add_user("name", "password")

    password, err:=users.get_users_password("gnome")
    if password!="" || err==nil{
        t.Error()
    }

    password, err=users.get_users_password("name")
    if password!="password" || err!=nil{
        t.Error()
    }
}

func TestUsersChange_password(t *testing.T){
    users:=new_users()
    users.add_user("name", "password")

    if users.change_password("gnome", "password", "otherpassword")==nil{
        t.Error()
    }

    if users.change_password("name", "otherpassword", "otherpassword")==nil{
        t.Error()
    }

    if users.change_password("name", "password", "")==nil{
        t.Error()
    }

    if users.users[0].Password!="password"{
        t.Error()
    }

    if users.change_password("name", "password", "otherpassword")!=nil{
        t.Error()
    }

    if users.users[0].Password!="otherpassword"{
        t.Error()
    }
}