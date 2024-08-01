// Code generated by sqlc-http (https://github.com/walterwanderley/sqlc-http).

package user

// NewService is a constructor of a interface { func RegisterHandlers(*http.ServeMux) } implementation.
// Use this function to customize the server by adding middlewares to it.
func NewService(querier *Queries) *CustomService {
	return &CustomService{Service: Service{querier: querier}}
}
