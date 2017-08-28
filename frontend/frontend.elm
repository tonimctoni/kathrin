import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Array

--import SHA exposing ( sha1bytes, sha224bytes, sha256bytes, sha1sum, sha224sum, sha256sum)

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






type alias Entries =
  { date: String
  , entries : Array.Array String
  }

type alias Model = {days_in_the_future : Int, entries : Entries, error : String}



init: (Model, Cmd Msg)
init = (Model 0 (Entries "" Array.empty) "", send_get_entries_request 0)



type Msg = NextDay | PreviousDay | EntriesArrived (Result Http.Error Entries)


nav_bar: Html Msg
nav_bar =
  nav [class "navbar navbar-inverse"]
  [ div [class "container-fluid"]
    [ div [class "navbar-header"]
      [ a [class "navbar-brand", href "/"] [text "Programs Name"]
      ]
    , ul [class "nav navbar-nav"]
      [ li [class "active"] [a [href "/"] [text "Plan"]]
      , li [] [a [href "/entries"] [text "My Entries"]]
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
    [ div [class "col-md-1"] [button [class left_button_class, onClick PreviousDay, style [("margin", ".25cm")]] [span [class "glyphicon glyphicon-menu-left"] [], text "Left"]]
    --, div [class "col-md-2"] []
    , div [class "col-md-3"] [h3 [style [("float", "centar")]] [text model.entries.date]]
    --, div [class "col-md-2"] []
    , div [class "col-md-1"] [button [class right_button_class, onClick NextDay, style [("margin", ".25cm"), ("float", "left")]] [text "Right", span [class "glyphicon glyphicon-menu-right"] []]]
    ]



user_row : Model -> Int -> Html Msg
user_row model i =
  let
    entry_row_free i =
      div [class "row"]
      [ p [class "bg-info text-white", style [("margin", ".1cm")]]
        [ p [] [text ((toString i)++" - "++(toString (i+1))++":")]
        --, text "_"
        , button [style [("margin-left", ".5cm")]] [span [class "glyphicon glyphicon-pencil"] []]
        ]
      ]

    entry_row_occ i name =
      div [class "row"]
      [ p [class "bg-success text-white", style [("margin", ".1cm")]]
        [ p [] [text ((toString i)++" - "++(toString (i+1))++":")]
        , strong [] [text name]
        , button [style [("margin-left", ".5cm")]] [span [class "glyphicon glyphicon-remove"] []]
        ]
      ]

    entry_row_error = div [class "row"] [p [class "bg-danger text-white", style [("margin", ".1cm")]] [text "ERROR"]]
  in
  case Array.get i model.entries.entries of
    Just "" -> entry_row_free i
    Just name -> entry_row_occ i name
    Nothing -> entry_row_error


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
  --, user_row model 24
  --, user_row model 25
  ]

view: Model -> Html Msg
view model =
  div []
  --[ node "link" [ rel "stylesheet", href "https://maxcdn.bootstrapcdn.com/bootstrap/3.3.7/css/bootstrap.min.css"] []
  [ node "link" [ rel "stylesheet", href "http://localhost:8000/bootstrap/css/bootstrap.min.css"] []
  , nav_bar
  , div [class "container", style [("background-color", "#D0D0D0"), ("border-radius", "6px")]]
    [ date_row model
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
      PreviousDay -> ({model | days_in_the_future=previous_day}, send_get_entries_request previous_day)
      NextDay -> ({model | days_in_the_future=next_day}, send_get_entries_request next_day)
      EntriesArrived (Ok entries) -> (Model model.days_in_the_future entries "", Cmd.none)
      EntriesArrived (Err err) -> case err of
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