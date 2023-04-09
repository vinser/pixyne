# Pixyne*

\**Created using [Fyne](https://github.com/fyne-io/fyne) GUI library*


Use Pixyne to quickly review your photo folder and safely delete bad and similar shots.  
The application has a simple and intuitive interface. With one click, you can move a photo to a trash subfolder.  
You may also set or correct the EXIF shooting date to the file date or to a manually entered date.  


![image](docs/pixyneapp.jpg)

When saving the changes it is possible to change all file names to the EXIF shooting date.

## Installation

The Pixine app doesn't require installation - just download the executable and run it.  
*Now it is ready only for the Windows platform - [download](https://github.com/vinser/pixyne/releases/download/v1.0.0/pixine.exe).*  

You can also install app directly from the source code using the Fyne command.  
To do so you will need to have Go and C compilers installed - *see the fyne [prerequisites](https://developer.fyne.io/started/).*  
Once set up execute the following:
```
go get fyne.io/fyne/v2/cmd/fyne
fyne get github.com/vinser/pixyne
```
Or you can download the code directly from git:
```
git clone https://github.com/vinser/pixyne.git
```
