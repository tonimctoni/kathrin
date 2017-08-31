import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick, onInput)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Array

send_get_entries_request: Int -> Cmd Msg
send_get_entries_request days_in_the_future =
  let
    body =
      [("days_in_the_future", Encode.int days_in_the_future)]
      |> Encode.object
      |> Http.jsonBody

    entries_decoder = Decode.map2 Entries
      (Decode.field "Date" Decode.string)
      (Decode.field "Entries" (Decode.array Decode.string))
  in
    Http.send EntriesArrived (Http.post "/get_entries" (body) entries_decoder)


send_add_entry_request: Model -> Cmd Msg
send_add_entry_request model =
  if model.name=="" then
    Cmd.none
  else if model.password=="" then
    Cmd.none
  else
    case model.active_entry of
      Nothing -> Cmd.none
      Just active_entry ->
        let
          body =
            [ ("days_in_the_future", Encode.int model.days_in_the_future)
            , ("date", Encode.string model.entries.date)
            , ("active_entry", Encode.int active_entry)
            , ("name", Encode.string model.name)
            , ("password", Encode.string model.password)
            ]
            |> Encode.object
            |> Http.jsonBody

          return_code_decoder = Decode.map ReturnCode
            (Decode.field "Return_code" Decode.int)
        in
          Http.send ReturnCodeArrived (Http.post "/add_entry" (body) return_code_decoder)


send_remove_entry_request: Model -> Cmd Msg
send_remove_entry_request model =
  if model.password=="" then
    Cmd.none
  else
    case model.active_entry of
      Nothing -> Cmd.none
      Just active_entry ->
        case Array.get active_entry model.entries.entries of
          Just "" -> Cmd.none
          Nothing -> Cmd.none
          Just name ->
            let
              body =
                [ ("days_in_the_future", Encode.int model.days_in_the_future)
                , ("date", Encode.string model.entries.date)
                , ("active_entry", Encode.int active_entry)
                , ("name", Encode.string name)
                , ("password", Encode.string model.password)
                ]
                |> Encode.object
                |> Http.jsonBody

              return_code_decoder = Decode.map ReturnCode
                (Decode.field "Return_code" Decode.int)
            in
              Http.send ReturnCodeArrived (Http.post "/remove_entry" (body) return_code_decoder)




type alias ReturnCode =
  { return_code: Int
  }

type alias Entries =
  { date: String
  , entries : Array.Array String
  }

type alias Model =
  { days_in_the_future: Int
  , entries: Entries
  , active_entry: Maybe Int
  , name: String
  , password: String
  , error: String
  , return_code: Int}



init: (Model, Cmd Msg)
init = (Model 0 (Entries "" Array.empty) Nothing "" "" "" 0, send_get_entries_request 0)



type Msg
  = NextDay
  | PreviousDay
  | EntriesArrived (Result Http.Error Entries)
  | ShowEntryForm (Maybe Int)
  | SetName String
  | SetPassword String
  | SendAddEntryRequest
  | ReturnCodeArrived (Result Http.Error ReturnCode)
  | SendRemoveEntryRequest


nav_bar: Html Msg
nav_bar =
  nav [class "navbar navbar-inverse"]
  [ div [class "container-fluid"]
    [ div [class "navbar-header"]
      [ a [class "navbar-brand", href "/"] [text "Programs Name"]
      ]
    , ul [class "nav navbar-nav"]
      [ li [class "active"] [a [href "/"] [text "Plan"]]
      , li [] [a [href "/change_password"] [text "Change Password"]]
      , li [class "dropdown"]
        [ a [class "dropdown-toggle", attribute "data-toggle" "dropdown", href "#"] [text "Admin"]
        ,  ul [class "dropdown-menu"]
          [ li [] [a [href "/see_all"] [text "See All"]]
          , li [] [a [href "/remove_old"] [text "Remove Old Entries"]]
          ]
        ]
      ]
    ]
  ]

date_row : Model -> Html Msg
date_row model =
  let
    left_button_class = if model.days_in_the_future>0 then "btn btn-default" else "btn btn-default disabled"
    right_button_class = "btn btn-default"
  in
    div [class "row"]
    [ div [class "col-md-1"] [button [class left_button_class, onClick PreviousDay, style [("margin", ".25cm")]] [span [class "glyphicon glyphicon-menu-left"] [], text "Last"]]
    --, div [class "col-md-2"] []
    , div [class "col-md-3"] [h3 [style [("float", "centar")]] [text model.entries.date]]
    --, div [class "col-md-2"] []
    , div [class "col-md-1"] [button [class right_button_class, onClick NextDay, style [("margin", ".25cm"), ("float", "left")]] [text "Next", span [class "glyphicon glyphicon-menu-right"] []]]
    ]




user_row : Model -> Int -> Html Msg
user_row model i =
  let
    add_entry_form: Model -> Html Msg
    add_entry_form model =
      div []
        [ input [type_ "text", placeholder "Name", onInput SetName] []
        , input [type_ "password", placeholder "Password", onInput SetPassword] []
        , button [disabled (if model.name=="" || model.password=="" then True else False), onClick SendAddEntryRequest] [text "OK"]
        ]
    remove_entry_form: Model -> Html Msg
    remove_entry_form model =
      div [] 
        [ input [type_ "password", placeholder "Password", onInput SetPassword] []
        , button [disabled (if model.password=="" then True else False)
        , onClick SendRemoveEntryRequest] [text "OK"]
        ]
  in
    case Array.get i model.entries.entries of
      Just "" -> 
        div [class "row"]
        [ p [class "bg-info text-white", style [("margin", ".1cm")]]
          [ p [] [text ((toString i)++" - "++(toString (i+1))++":")]
          , button [style [("margin-left", ".5cm")], onClick (if (model.active_entry==Just i) then (ShowEntryForm Nothing) else (ShowEntryForm (Just i)))] [span [class "glyphicon glyphicon-pencil"] []]
          , if model.active_entry==(Just i) then (add_entry_form model) else (div [] [])
          ]
        ]
      Just name ->
        div [class "row"]
        [ p [class "bg-success text-white", style [("margin", ".1cm")]]
          [ p [] [text ((toString i)++" - "++(toString (i+1))++":")]
          , strong [] [text name]
          , button [style [("margin-left", ".5cm")], onClick (if (model.active_entry==Just i) then (ShowEntryForm Nothing) else (ShowEntryForm (Just i)))] [span [class "glyphicon glyphicon-remove"] []]
          , if model.active_entry==(Just i) then (remove_entry_form model) else (div [] [])
          ]
        ]
      Nothing ->
        div [class "row"] [p [class "bg-danger text-white", style [("margin", ".1cm")]] [text "ERROR"]]


user_rows : Model -> Html Msg
user_rows model =
  div []
  [ user_row model 0
  , user_row model 1
  , user_row model 2
  , user_row model 3
  , user_row model 4
  , user_row model 5
  , user_row model 6
  , user_row model 7
  , user_row model 8
  , user_row model 9
  , user_row model 10
  , user_row model 11
  , user_row model 12
  , user_row model 13
  , user_row model 14
  , user_row model 15
  , user_row model 16
  , user_row model 17
  , user_row model 18
  , user_row model 19
  , user_row model 20
  , user_row model 21
  , user_row model 22
  , user_row model 23
  ]

error_message: Model -> Html Msg
error_message model =
  if model.error=="" then
    div [] []
  else
    div [class "row alert alert-danger"] 
    [ strong [] [text ("Error! ("++model.error++")")]
    ]

return_code_message: Model -> Html Msg
return_code_message model =
  case model.return_code of
    0 -> div [] []
    1 -> div [class "row alert alert-danger"] [strong [] [text ("Error! (Username does not exist)")]]
    2 -> div [class "row alert alert-danger"] [strong [] [text ("Error! (Wrong password)")]]
    3 -> div [class "row alert alert-danger"] [strong [] [text ("Error! (Inconsistent date)")]]
    4 -> div [class "row alert alert-danger"] [strong [] [text ("Error! (Reservation already exists)")]]

    20 -> div [class "row alert alert-success"] [strong [] [text ("Success! (Entry added)")]]
    21 -> div [class "row alert alert-success"] [strong [] [text ("Success! (Entry removed)")]]
    _ -> div [] []

view: Model -> Html Msg
view model =
  div []
  [ node "link" [ rel "stylesheet", href "/bootstrap/css/bootstrap.min.css"] []
  , node "script" [ src "/bootstrap/js/jquery.min.js"] []
  , node "script" [ src "/bootstrap/js/bootstrap.min.js"] []
  , nav_bar
  , div [class "container", style [("background-color", "#D0D0D0"), ("border-radius", "6px")]]
    [ date_row model
    , error_message model
    , return_code_message model
    , user_rows model
    ]
  ]



update: Msg -> Model -> (Model, Cmd Msg)
update msg model =
  let
    previous_day = if model.days_in_the_future>0 then model.days_in_the_future-1 else 0
    next_day = model.days_in_the_future+1
  in
    case msg of
      SendRemoveEntryRequest -> (model, send_remove_entry_request model)
      SendAddEntryRequest -> (model, send_add_entry_request model)
      SetName name -> ({model | name=name}, Cmd.none)
      SetPassword password -> ({model | password=password}, Cmd.none)
      ShowEntryForm active_entry -> ({model | active_entry=active_entry, name="", password=""}, Cmd.none)
      PreviousDay -> ({model | days_in_the_future=previous_day, return_code=0}, send_get_entries_request previous_day)
      NextDay -> ({model | days_in_the_future=next_day, return_code=0}, send_get_entries_request next_day)
      EntriesArrived (Ok entries) -> ({model | entries=entries, name="", password="", active_entry=Nothing, error=""}, Cmd.none)
      EntriesArrived (Err err) -> case err of
        Http.BadUrl s -> ({model | error="BadUrl: "++s}, Cmd.none)
        Http.Timeout -> ({model | error="Timeout"}, Cmd.none)
        Http.NetworkError -> ({model | error="NetworkError"}, Cmd.none)
        Http.BadStatus _ -> ({model | error="BadStatus"}, Cmd.none)
        Http.BadPayload s _ -> ({model | error="BadPayload: "++s}, Cmd.none)
      ReturnCodeArrived (Ok return_code) -> ({model | return_code=return_code.return_code}, send_get_entries_request model.days_in_the_future)
      ReturnCodeArrived (Err err) -> case err of
        Http.BadUrl s -> ({model | error="BadUrl: "++s}, Cmd.none)
        Http.Timeout -> ({model | error="Timeout"}, Cmd.none)
        Http.NetworkError -> ({model | error="NetworkError"}, Cmd.none)
        Http.BadStatus _ -> ({model | error="BadStatus"}, Cmd.none)
        Http.BadPayload s _ -> ({model | error="BadPayload: "++s}, Cmd.none)



subscriptions: Model -> Sub Msg
subscriptions model=
    Sub.none



main: Program Never Model Msg
main =
    program
        {init=init
        ,view=view
        ,update=update
        ,subscriptions=subscriptions
        }