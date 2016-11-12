module Main exposing (..)

import Html exposing (Html)
import Html.App as App
import Html.Attributes as Attr
import Books
import Http

type alias Flags = { loggedIn : Bool }

main : Program Flags
main = App.programWithFlags { init = init
                            , update = update
                            , view = view
                            , subscriptions = subscriptions
                            }


type alias Model = { loggedIn : Bool, books : List Books.Book }

init : Flags -> (Model, Cmd Msg)
init flags = ({ loggedIn = flags.loggedIn, books = [] }, Books.get BooksFailed BooksSucceeded)

type Msg = NoOp | BooksFailed Http.Error | BooksSucceeded (List Books.Book)

update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
    case msg of
        NoOp -> (model, Cmd.none)

        BooksFailed error ->
            (model, Cmd.none)

        BooksSucceeded books ->
            ({model | books = books}, Cmd.none)

view : Model -> Html Msg
view model =
    let
        login = [ Html.div [ Attr.class "cover" ]
                      [ Html.a [ Attr.href "/sign-in" ] [ Html.text "Sign-in" ]
                      ]
                ]

        table = [ Html.input [ Attr.id "filter"
                             , Attr.name "filter"
                             , Attr.type' "text"
                             , Attr.placeholder "Search" ] []
                , Html.table []
                    [ Html.thead []
                          [ Html.tr []
                                [ Html.th [] [ Html.text "Title" ]
                                , Html.th [] [ Html.text "Author" ]
                                , Html.th [] [ Html.text "Added" ]
                                , Html.th [] [ Html.text "Editions" ]
                                ]
                          ]
                    , Html.tbody [] []
                    ]
                ]
    in
        Html.div [] <|
             [ Html.h1 []
                   [ Html.text "alexandria" ]
             ] ++ (if model.loggedIn then table else login)


subscriptions : Model -> Sub Msg
subscriptions model = Sub.none
