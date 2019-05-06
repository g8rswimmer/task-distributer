# Task Distributer

## How to start

This application is hosted by [Heroku](https://www.heroku.com) and uses a [Postgres](https://www.postgresql.org/) database.  It is recommended to have the [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) installed which would allow you to run locally or on your own dyno.  However, this application is running on a dyno, https://ancient-mountain-96195.herokuapp.com/.

### Running Locally
After installing the Heroku CLI, you can run the application locally.
```
heroku local -e .env.test
```
This will run the application using the `Postgres` database on port `5000`.  Please note, the before testing you may need to make sure that all of the tasks are completed.

## Limitations

There are some limitations that should be noted:
* Unit Tests
  - Testing was done via `curl` and `postman`, however testing was limited due to mocking of the database (`go-sqlmock` was being considered but was unable to get it to work properly)
   
## APIs

The following are the `APIs` that are currently supported.

### Create Task
This `API` will accept a task and attempt to distribute it to an available agent.

#### URI

`/v1/task/create`

#### Content Type

JSON

#### HTTP Method
POST

#### Parameters
None.

#### Request Body
| Field    | Required | Type             | Description                                                         |
|----------|----------|------------------|---------------------------------------------------------------------|
| name     | yes      | string           | The name of the task                                                    |
| skills   | yes      | array of strings | An array of skills required by the task.  Accepted skills are skill1, skill2, and skill3 |
| priority | yes      | string           | The priority of the task.  Accepted priorities are low and high.    |

```
{
	"name": "Test Name",
	"skills": ["skill1"],
	"priority": "high"
}
```
#### Response Body
| Field   | Type   | Description                                                  |
|---------|--------|--------------------------------------------------------------|
| success | bool   | If the application was successfully distributed to an agent. |
| task    | object | Description of the task with the assigned agent. Only present if success is true |
| error_message    | string | A description of the error that occured.  Only present if sucess is false |

##### Task
| Field         | Type             | Description                                                                    |
|---------------|------------------|--------------------------------------------------------------------------------|
| id            | string           | UUID of the task.  Can be used to access the status and setting of parameters. |
| name          | string           | The name of the task                                                           |
| skills        | array of strings | An array of skills required by the task.                                       |
| priority      | string           | The priority of the task.                                                      |
| start_time    | Date and time    | The date and time of when the task has been distributed to an agent            |
| status        | string           | The status of the task, currently set to assigned.                             |
| complete_time | Date and time    | The date and time of when the task was completed by the agent                  |
| agent         | string           | The UUID of the agent assigned to the task                                     |

#### Examples
 ```
 curl -d '{"name": "Task Test","skills": ["skill1"],"priority": "low"}' -H "Content-Type: application/json" -X POST https://ancient-mountain-96195.herokuapp.com/v1/task/create
 ```
##### Success
```
{
    "success": true,
    "task": {
        "id": "bj7rmmrk7c874r7vb8ng",
        "name": "Test Name",
        "skills": [
            "skill1"
        ],
        "priority": "high",
        "status": "Assigned",
        "start_time": "2019-05-06T04:43:07.143378962Z",
        "complete_time": "0001-01-01T00:00:00Z",
        "assigned_agent": "1000"
    }
}
```
##### Errors
```
{
    "success":false,
    "error_message":"unable to find an agent to assign the task"
}
```

```
{
    "success":false,
    "error_message":"Invalid priority task priority is not supported zzz"
}
```

```
{
    "success":false,
    "error_message":"Required field missing name field must be present"
}
```
### Task Status

#### URI

#### Format

#### HTTP Method

#### Parameters

#### Reuest Body

#### Response Body

#### Example 

### Task Complete

#### URI

#### Format

#### HTTP Method

#### Parameters

#### Reuest Body

#### Response Body

#### Example 

### Agent List

#### URI

#### Format

#### HTTP Method

#### Parameters

#### Reuest Body

#### Response Body

#### Example 
