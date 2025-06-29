' DreamHub Launcher 启动脚本
' 使用 VBScript 在后台启动程序，避免显示命令行窗口

Set WshShell = CreateObject("WScript.Shell")
Set objFSO = CreateObject("Scripting.FileSystemObject")

' 获取脚本所在目录
strPath = objFSO.GetParentFolderName(WScript.ScriptFullName)

' 切换到程序目录
WshShell.CurrentDirectory = strPath

' 在后台启动 dreamhub.exe
WshShell.Run "dreamhub.exe", 0, False

' 显示提示信息
MsgBox "DreamHub Launcher 已启动！" & vbCrLf & vbCrLf & _
       "程序正在系统托盘运行。" & vbCrLf & _
       "请查看屏幕右下角的托盘图标。", _
       vbInformation, "DreamHub"