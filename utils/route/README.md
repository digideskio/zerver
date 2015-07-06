#### route

### Example
```Go
import "log"
import "github.com/cosiner/zerver/util/route"

func init() {
    var routes = route.
        Get("/user/info", handler1).
        Post("/user/article", handler2).
        Delete("/user/article/:id", handler3)
        ...
    
    if err := routes.Apply(Server.Router); err != nil {
        log.Panicln(err)
    }
}
```
