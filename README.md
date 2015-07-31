# HeartRating
### 

## API
    POST /api/save
    expected body:

     {
        "heart-score": 3,      # calculated score from Android app
        "watch-time": 1100000, # milliseconds watched
        "title": "Daredevil",  # title name
        "show": "Pilot",       # tv show title
        "user": "Bob"          # user account
     }

    GET  /api/sessions/:user
    response:

    {
    "user": {
        "Id": 1,
        "Name": "alice"
    },
    "sessions": [
        {
            "Title": "title",
            "Show": "show",
            "Heart": 1,
            "Duration": 1
        },
        {
            "Title": "title",
            "Show": "show",
            "Heart": 1,
            "Duration": 1
        },
        {
            "Title": "title",
            "Show": "show",
            "Heart": 1,
            "Duration": 1
        }
    ],
    "status": "success"
    }

    GET  /api/users
    response

    {
        "users": [
            {
                "Id": 11,
                "Name": "Kevin"
            },
            {
                "Id": 6,
                "Name": "George"
            }
        ],
        "status": "success"
    }
 
