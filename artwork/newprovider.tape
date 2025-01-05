Output artwork/newprovider.gif

Require echo
Require go

Set Shell "zsh"
Set CursorBlink false
Set FontSize 32
Set Width 2000
Set Height 1000
Set TypingSpeed 0.2

Set Shell zsh
Sleep 500ms
Type "go build -o gmsg"
Sleep 700ms
Enter

Sleep 3s

Type "./gmsg newprovider"
Enter

Sleep 1s

# Select silicon, down 11 times
Down 2
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Down
Sleep 100ms
Enter

# Use default api base
Sleep 3s
Enter
Sleep 2s
# Type api key, random key
Type sk-123ywgrekwchqtkhvxgzeeuqairyzxzgcpssuzczorltitsi
Sleep 2s
Enter

# Use default max token
Sleep 2s
Enter
# Use default model
Sleep 2s
Enter

# Overwrite
Sleep 3s
Type y
Enter

Sleep 9s