package main

import (
    "github.com/gorilla/mux"
    "github.com/justinas/alice"
    "log"
    "net/http"
    "time"
    "github.com/go-ego/riot"
    "github.com/go-ego/riot/types"
    "strconv"
)


var searcher = riot.Engine{}

func main() {
    errorChain := alice.New(loggerHandler, recoverHandler)

    var r = mux.NewRouter()
    r.HandleFunc("/", rootHandler)
    r.HandleFunc("/search", searchHandler).Name("search")
    r.HandleFunc("/index", indexHandler).Name("index")


    http.Handle("/", errorChain.Then(r))

    server := &http.Server{
        Addr: ":8881",
        Handler: r,
    }

    initSearcher()

    log.Printf("Service UP\n")

    err := server.ListenAndServe()
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }

}

func initSearcher() {
    searcher.Init(types.EngineOpts{
        IndexerOpts: &types.IndexerOpts{
            IndexType: types.DocIdsIndex,
        },
        UseStorage:    true,
        StorageFolder: "./riot-index",
    })

    text := "Google Is Experimenting With Virtual Reality Advertising"
    text1 := `Google accidentally pushed Bluetooth google update for Home speaker early`
    text2 := `Google is testing another Search results layout with rounded cards, new colors, and the 4 mysterious colored dots again`
    text3 := "Google testing text search"
    text4 := "Google testing search"

    // Add the document to the index, docId starts at 1
    searcher.IndexDoc(1, types.DocIndexData{Content: text})
    searcher.IndexDoc(2, types.DocIndexData{Content: text1})
    searcher.IndexDoc(3, types.DocIndexData{Content: text2})
    searcher.IndexDoc(4, types.DocIndexData{Content: text3})
    searcher.IndexDoc(5, types.DocIndexData{Content: text4})

    searcher.Flush()

    log.Printf("Search flush\n")
}


func searchHandler(w http.ResponseWriter, r *http.Request) {
    k := r.URL.Query().Get("keyword")

    search := searcher.Search(types.SearchReq{Text: k})
    log.Println("search...", search)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    id, _ := strconv.ParseUint(r.URL.Query().Get("id"), 10, 64)
    content := r.URL.Query().Get("content")

    searcher.IndexDoc(id, types.DocIndexData{Content: content})
    log.Println("index %d", id)
}



func rootHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("tadam")
}


func loggerHandler(h http.Handler) http.Handler {

    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        h.ServeHTTP(w, r)
        log.Printf("<< %s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

func recoverHandler(next http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %+v", err)
                http.Error(w, http.StatusText(500), 500)
            }
        }()

        next.ServeHTTP(w, r)
    }

    return http.HandlerFunc(fn)
}
