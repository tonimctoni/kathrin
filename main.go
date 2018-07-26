package main;

import "encoding/json"
import "net/http"
import "fmt"
import "os"
import "time"
import "strings"

func add_file_to_mux(mux *http.ServeMux, filepath string, mimetype string){
    mux.HandleFunc(fmt.Sprintf("/%s", filepath), func (w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", mimetype)
        http.ServeFile(w, r, filepath)
    })
}

func add_file_to_mux_at_path(mux *http.ServeMux, webpath string, filepath string, mimetype string){
    mux.HandleFunc(webpath, func (w http.ResponseWriter, r *http.Request){
        w.Header().Set("Content-Type", mimetype)
        http.ServeFile(w, r, filepath)
    })
}

func (u *Users) http_get_entries(w http.ResponseWriter, r *http.Request){
    var to_get struct{
        Days_in_the_future int
    }

    var to_send struct{
        Date string
        Entries [24] string
    }

    if r.Method!="POST"{
        http.Error(w, "Request to this address must be POST.", http.StatusMethodNotAllowed)
    }

    err:=json.NewDecoder(r.Body).Decode(&to_get)
    if err!=nil{
        http.Error(w, "Request's data could not be parsed.", http.StatusBadRequest)
        return
    }

    // Get entries for the specified day
    now:=time.Now().AddDate(0,0,to_get.Days_in_the_future)
    year, month, day:=now.Date()
    weekday:=now.Weekday()
    to_send.Date=fmt.Sprintf("%s, %d of %s", weekday.String(), day, month.String())
    entries:=u.get_entries_on_day(Entry{year, int(month), day, 0})
    copy(to_send.Entries[:], entries[:])

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&to_send)
}

func (u *Users) http_add_entry(w http.ResponseWriter, r *http.Request){
    var to_get struct{
        Days_in_the_future int
        Date string
        Active_entry int
        Name string
        Password string
    }

    var to_send struct{
        Return_code int
    }

    if r.Method!="POST"{
        http.Error(w, "Request to this address must be POST.", http.StatusMethodNotAllowed)
    }

    err:=json.NewDecoder(r.Body).Decode(&to_get)
    if err!=nil{
        http.Error(w, "Request's data could not be parsed.", http.StatusBadRequest)
        return
    }

    // Get password, on error send error code
    password, err:=u.get_users_password(to_get.Name)
    if err!=nil{
        to_send.Return_code=1
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // See if password is correct, on error send error code
    if to_get.Password!=password{
        to_send.Return_code=2
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // See if dates are inconsistent
    now:=time.Now().AddDate(0,0,to_get.Days_in_the_future)
    year, month, day:=now.Date()
    weekday:=now.Weekday()
    if fmt.Sprintf("%s, %d of %s", weekday.String(), day, month.String())!=to_get.Date{
        to_send.Return_code=3
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // Try to add the actual entry
    new_entry:=Entry{year, int(month), day, to_get.Active_entry}
    err=u.add_entry(to_get.Name, new_entry)
    if err!=nil{
        to_send.Return_code=4
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // Save changes to file
    err=u.to_file("users.json")
    if err!=nil{
        fmt.Fprintln(os.Stderr, err)
    }
    fmt.Println("Entry added:", to_get.Name, new_entry.String())

    // If the program got here, the reservation was added correctly. Send good return code
    to_send.Return_code=20
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&to_send)
    return
}

func (u *Users) http_remove_entry(w http.ResponseWriter, r *http.Request){
    var to_get struct{
        Days_in_the_future int
        Date string
        Active_entry int
        Name string
        Password string
    }

    var to_send struct{
        Return_code int
    }

    if r.Method!="POST"{
        http.Error(w, "Request to this address must be POST.", http.StatusMethodNotAllowed)
    }

    err:=json.NewDecoder(r.Body).Decode(&to_get)
    if err!=nil{
        http.Error(w, "Request's data could not be parsed.", http.StatusBadRequest)
        return
    }

    // Get password, on error send error code
    password, err:=u.get_users_password(to_get.Name)
    if err!=nil{
        to_send.Return_code=1
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // See if password is correct, on error send error code
    if to_get.Password!=password{
        to_send.Return_code=2
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // See if dates are inconsistent
    now:=time.Now().AddDate(0,0,to_get.Days_in_the_future)
    year, month, day:=now.Date()
    weekday:=now.Weekday()
    if fmt.Sprintf("%s, %d of %s", weekday.String(), day, month.String())!=to_get.Date{
        to_send.Return_code=3
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // Try to remove the actual entry
    entry_to_remove:=Entry{year, int(month), day, to_get.Active_entry}
    err=u.remove_entry(to_get.Name, entry_to_remove)
    if err!=nil{
        to_send.Return_code=4
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(&to_send)
        return
    }

    // Save changes to file
    err=u.to_file("users.json")
    if err!=nil{
        fmt.Fprintln(os.Stderr, err)
    }
    fmt.Println("Entry removed:", to_get.Name, entry_to_remove.String())

    // If the program got here, the reservation was added correctly. Send good return code
    to_send.Return_code=20
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&to_send)
    return
}

func (u *Users) http_change_password(w http.ResponseWriter, r *http.Request){
    if r.Method=="GET"{
        w.Header().Set("Content-Type", "text/html")
        http.ServeFile(w, r, "frontend/change_password.html")
        return
    }

    if r.Method!="POST"{
        http.Error(w, "Request to this address must be GET or POST.", http.StatusMethodNotAllowed)
    }

    // Get form data
    name:=r.FormValue("name")
    password:=r.FormValue("password")
    new_password1:=r.FormValue("new_password1")
    new_password2:=r.FormValue("new_password2")
    character_whitelist:="abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    min_password_length:=4
    max_password_length:=32

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
    err:=u.change_password(name, password, new_password1)
    if err!=nil{
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Save changes to file
    err=u.to_file("users.json")
    if err!=nil{
        fmt.Fprintln(os.Stderr, err)
    }
    fmt.Println("Password changed by user:", name)

    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte("Password changed successfully"))
    return
}

func (u *Users) http_see_all(w http.ResponseWriter, r *http.Request){
    if r.Method=="GET"{
        w.Header().Set("Content-Type", "text/html")
        http.ServeFile(w, r, "frontend/admin.html")
        return
    }

    if r.Method!="POST"{
        http.Error(w, "Request to this address must be GET or POST.", http.StatusMethodNotAllowed)
    }

    // Get form data
    admin_password:=r.FormValue("admin_password")
    admin_password_, err:=u.get_users_password("admin")
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
    json_users,err:=u.as_json()
    if err!=nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(json_users)
    return
}

func (u *Users) http_remove_old(w http.ResponseWriter, r *http.Request){
    if r.Method=="GET"{
        w.Header().Set("Content-Type", "text/html")
        http.ServeFile(w, r, "frontend/admin.html")
        return
    }

    if r.Method!="POST"{
        http.Error(w, "Request to this address must be GET or POST.", http.StatusMethodNotAllowed)
    }

    // Get form data
    admin_password:=r.FormValue("admin_password")
    admin_password_, err:=u.get_users_password("admin")
    if err!=nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if admin_password_!=admin_password{
        http.Error(w, "Wrong password", http.StatusUnauthorized)
        return
    }

    u.remove_old_entries()
    // Save changes to file
    err=u.to_file("users.json")
    if err!=nil{
        fmt.Fprintln(os.Stderr, err)
    }
    fmt.Println("Old entries removed")
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte("Old entries removed"))
    return
}


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
    // users.to_file("users.json")
    // return

    users, err:=from_file("users.json")
    if err!=nil{
        fmt.Fprintln(os.Stderr, err)
    }
    users.add_user("admin", "password")
    users.Sort()


    mux:=http.NewServeMux()
    add_file_to_mux_at_path(mux, "/", "frontend/index.html", "text/html")
    add_file_to_mux(mux, "bootstrap/css/bootstrap.min.css", "text/css")
    add_file_to_mux(mux, "bootstrap/js/bootstrap.min.js", "application/javascript")
    add_file_to_mux(mux, "bootstrap/js/jquery.min.js", "application/javascript")
    add_file_to_mux(mux, "bootstrap/fonts/glyphicons-halflings-regular.ttf", "text/plain")
    add_file_to_mux(mux, "bootstrap/fonts/glyphicons-halflings-regular.woff", "text/plain")
    add_file_to_mux(mux, "bootstrap/fonts/glyphicons-halflings-regular.woff2", "text/plain")
    mux.HandleFunc("/get_entries", users.http_get_entries)
    mux.HandleFunc("/add_entry", users.http_add_entry)
    mux.HandleFunc("/remove_entry", users.http_remove_entry)
    mux.HandleFunc("/change_password", users.http_change_password)
    mux.HandleFunc("/see_all", users.http_see_all)
    mux.HandleFunc("/remove_old", users.http_remove_old)


    if err:=http.ListenAndServe(":8000", mux);err!=nil{
        fmt.Fprintln(os.Stderr, err)
    }
}
