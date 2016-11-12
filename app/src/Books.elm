module Books exposing (Book
                      , get)

import Json.Decode as Decode
import Json.Decode.Pipeline as Pipeline
import Task exposing (Task)
import Http
import Dict exposing (Dict)

type alias Book = { id : String
                  , title : String
                  , author : String
                  , added : String
                  , editions : List Edition
                  , links : Dict String Href
                  }

booksDecoder : Decode.Decoder (List Book)
booksDecoder = Decode.list bookDecoder

bookDecoder : Decode.Decoder Book
bookDecoder =
    Pipeline.decode Book
        |> Pipeline.required "id" Decode.string
        |> Pipeline.required "title" Decode.string
        |> Pipeline.required "author" Decode.string
        |> Pipeline.required "added" Decode.string
        |> Pipeline.required "editions" editionsDecoder
        |> Pipeline.required "links" linksDecoder

type alias Edition = { id : String
                     , name : String
                     , links : Dict String Href
                     }

editionsDecoder : Decode.Decoder (List Edition)
editionsDecoder = Decode.list editionDecoder

editionDecoder : Decode.Decoder Edition
editionDecoder =
    Pipeline.decode Edition
        |> Pipeline.required "id" Decode.string
        |> Pipeline.required "name" Decode.string
        |> Pipeline.required "links" linksDecoder

type alias Href = { href : String
                  }

linksDecoder : Decode.Decoder (Dict String Href)
linksDecoder = Decode.dict linkDecoder

linkDecoder : Decode.Decoder Href
linkDecoder =
    Pipeline.decode Href
        |> Pipeline.required "href" Decode.string

get : (Http.Error -> msg) -> ((List Book) -> msg) -> Cmd msg
get err ok =
    let
        request = { verb = "GET"
                  , headers = [ ("Accept", "application/json") ]
                  , url = "http://localhost:8080/books"
                  , body = Http.empty
                  }
    in
        Task.perform err ok <|
            Http.fromJson booksDecoder (Http.send Http.defaultSettings request)
