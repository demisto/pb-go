Pandora Bots API for Golang
============================
Implementation of the public API as specified [here](https://developer.pandorabots.com/docs).


Implemented features
====================

| *Resource*                                    | *Description*                                | *Supported* |
|-----------------------------------------------|------------------------------------------------------------|
| GET /bot/APP_ID                               | List of bots                                 |true         |
| PUT /bot/APP_ID/BOTNAME                       | Create a bot                                 |true         |
| DELETE /bot/APP_ID/BOTNAME                    | Delete a bot                                 |true         |
| GET bot/APP_ID/BOTNAME                        | List of bot files                            |true         |
| PUT /bot/APP_ID/BOTNAME/FILE-KIND/FILENAME    | Upload a personality file                    |true         |
| PUT /bot/APP_ID/BOTNAME/properties            | Upload a properties/pdefaults file           |true         |
| DELETE /bot/APP_ID/BOTNAME/FILE-KIND/FILENAME | Delete  personality file                     |true         |
| DELETE /bot/APP_ID/BOTNAME/FILE-KIND          | Delete a properties/pdefaults file           |true         |
| GET /bot/APP_ID/BOTNAME/FILE-KIND/FILENAME    | Retrieve  personality file                   |true         |
| GET /bot/APP_ID/BOTNAME/FILE-KIND             | Retrieve a properties/pdefaults file         |true         |
| GET /bot/APP_ID/BOTNAME/verify                | Verify / compile a bot                       |true         |
| POST /talk/APP_ID/BOTNAME                     | Talk with a bot (including debug parameters) |true         |


Install
=======

If you have a go workplace setup and working you can simply do:

 ```go get github.com/demisto/pb-go```

 ```go install github.com/demisto/pb-go```


Usage
=====

In order how to use the `pb-go` module please have a look at the documentation and the `cli` directory.
You need to have a PandoraBots application id and user key to start using the API.
