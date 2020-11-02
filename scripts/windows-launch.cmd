@ECHO OFF

SETLOCAL

SET "FERR_EXE=%~dp0\bin\ferr.bin.exe"
SET "PATH=%~dp0\bin;%PATH%"

CALL %FERR_EXE% %*
EXIT /B %ERRORLEVEL%
