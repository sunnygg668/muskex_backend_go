@echo off
setlocal

set "filepath=%~1"
if "%filepath%"=="" set "filepath=proto.proto"

set "destdir=%~2"
if "%destdir%"=="" set "destdir=."

echo gen proto source for %filepath% to %destdir%

set dirname=%~dp1

REM 执行 protoc 命令
"D:/www/protoc-28.2-win64/bin/protoc.exe" -I %destdir% -I . -I "%GOPATH%\src" ^
    --go_out=%destdir%  --go_opt paths=source_relative ^
    --connect-go_out=%destdir%  --go_opt paths=source_relative ^
    --go-grpc_out=%destdir% --go-grpc_opt paths=source_relative ^
    --grpc-gateway_out=%destdir% --grpc-gateway_opt logtostderr=true,paths=source_relative,generate_unbound_methods=true,grpc_api_configuration=.\gen_gw.yaml ^
    --openapiv2_out=.\swagger --openapiv2_opt allow_merge=true,merge_file_name=foo,enums_as_ints=true,logtostderr=true,generate_unbound_methods=true ^
    %filepath%

set res=%ERRORLEVEL%
if %res% neq 0 (
    echo error gen protoc with status: %res%
    exit /b %res%
) else (
    echo finish gen %filepath% go-source to %destdir%
)

endlocal
exit /b 0
