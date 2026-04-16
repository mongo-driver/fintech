$ErrorActionPreference = "Stop"

go test `
  ./services/api-gateway/internal/app `
  ./services/auth-service/internal/app `
  ./services/user-service/internal/app `
  ./services/wallet-service/internal/app `
  ./services/notification-service/internal/app `
  ./shared/config `
  ./shared/security `
  ./shared/middleware `
  ./shared/grpcx `
  -coverprofile coverage.out `
  -covermode atomic
go tool cover -func coverage.out
