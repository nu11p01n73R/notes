# Notes

````
                 _ __   ___ | |_ ___  ___
                | '_ \ / _ \| __/ _ \/ __|
                | | | | (_) | ||  __/\__ \
                |_| |_|\___/ \__\___||___/
```

`notes` is a simple note making app written in golang.
Take notes in markdown with ease on you favourite editor.

# Commands

- `add` Add a new note.
- `remove notefile` Remove a note.
- `edit notefile` Edit a note.
- `search` Use [`fuz`](https://github.com/nu11p01n73R/fuz) to search the notes directory.

# Why markdown?

The main reason why markdown is used because of its rich syntax.
Also it allows you to publish your old notes on your favourite
bloging platform with just a copy-paste.

# Template

```
[TITLE]

[TAGS]

[CONTENT]

```
The notes will be identified using the title.


# Internals

The notes are saved at `.notes` directory in the users
home. Two new directories are created inside the root,

```
data
tags
```

The notes are organised based on the date it is created.
A note with title, `Hello World` created on 4th November
will be identified as,

```
20171104/hello_world.md
```

The directory structure for such a file will be,

```
data
    20171104
        new_title.md
```
For each tag in the note file, the note name will be
added in the corresponding file in `tags` directory.



# Options

```
-nd     Change the default root directory
```
