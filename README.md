# Pixyne

### Use Pixyne to quickly review your photo folder, safely delete bad and similar shots and fix shooting date without any AI.  

---

![image](docs/pixyneapp.jpg)

![image](docs/pixyneapp2.jpg)

## How to use

The application has a simple and intuitive interface, but there are some non-obvious things.   

With a click on photo you can mark to drop it in the trash.

When you are in a list view, clicking on a list row opens the corresponding photo.  

You may also set or correct the EXIF shooting date to the file date or to a manually entered date.

When you save changes, you can change all file names to EXIF shooting date format.  

The application stores the current state of the folder so you can undo changes at any time. But the changes will not actually be applied until you save. This is useful when working with a large number of photos and when you need to close the application in order to continue working later. If you open another folder, all unsaved changes will be lost.    

## Installation

The Pixine app doesn't require installation - just download the executable and run it.  
*The executable is currently ready for the Windows platform only - [download](https://github.com/vinser/pixyne/releases/download/v1.0.0/pixine.exe).*  

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
---
*Created using [Fyne](https://github.com/fyne-io/fyne) GUI library*  
*App icon designed by [Icon8](https://icon8.com)*  

