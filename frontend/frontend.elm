import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onClick)
import Http
import Json.Decode as Decode
import Json.Encode as Encode
import Array

import SHA exposing ( sha1bytes, sha224bytes, sha256bytes, sha1sum, sha224sum, sha256sum)

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
      , li [] [a [href "/reservations"] [text "My Reservations"]]
      ]
    ]
  ]

date_col : Model -> Html Msg
date_col model = 
  let
    left_button_class = if model.days_in_the_future>0 then "btn btn-default" else "btn btn-default disabled"
    right_button_class = "btn btn-default"
  in
    div [class "row"]
    [ div [class "col-md-1"] [button [class left_button_class, onClick PreviousDay] [text "Left"]]
    , div [class "col-md-10"] [h2 [style [("float", "center")]] [text model.entries.date]]
    , div [class "col-md-1"] [button [class right_button_class, onClick NextDay] [text "Right"]]
    ]

view: Model -> Html Msg
view model =
  div []
  [ node "link" [ rel "stylesheet", href "/bootstrap/css/bootstrap.min.css"] []
  , nav_bar
  , div [class "container", style [("background-color", "#EFFFEF"), ("border-radius", "6px")]]
    [ date_col model
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