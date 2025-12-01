# BackWater

A lightweight, CLI-based integration testing tool written in Go. It executes sequential HTTP requests defined in a JSON configuration file, supporting dynamic variable substitution, request chaining, and response validation.

## 🚀 Features

  * **JSON-Based Configuration**: Define test suites entirely in standard JSON.
  * **Variable Substitution**: Dynamically inject values into URLs, Headers, and Bodies using `$variable$` syntax.
  * **Request Chaining**: Extract data from a response (e.g., Auth Tokens, IDs) and store them as variables for subsequent tests.
  * **Global Variables**: Define environment-specific variables (like base URLs or secrets) once.
  * **Zero External Dependencies**: Built using Go standard library.

## 🛠 Installation & Build

Ensure you have [Go](https://go.dev/) installed.

1.  **Clone the repository**
2.  **Build the binary**:
    ```bash
    go build -o tester .
    ```

## 📖 Usage

Run the tool by pointing it to your test configuration file.

```bash
./tester --path ./path/to/your-test.json
```

**Flags:**

  * `--path`: (Optional) Path to the test JSON file. Defaults to `./test.json`.

-----

## 📄 Configuration Structure

The test file is a JSON object with the following structure:

| Field | Type | Description |
| :--- | :--- | :--- |
| `name` | `string` | The name of the test suite. |
| `variables` | `object` | Key-value pairs for global variables (e.g., API keys, hostnames). |
| `tests` | `array` | An ordered list of test cases to execute. |

### Test Case Definition (`tests` array)

Each object in the `tests` array represents a single HTTP step:

```json
{
  "num": 1,
  "method": "POST",
  "url": "http://localhost:8080/api/resource",
  "header": { "Content-Type": "application/json" },
  "body": { "key": "value" },
  "expected_status": "201 Created",
  "expected_response": {},
  "var_to_store": {
    "saved_id": "data.id"
  }
}
```

| Field | Description |
| :--- | :--- |
| `method` | HTTP method (GET, POST, PUT, DELETE, etc.). |
| `url` | Target URL. Supports variable substitution (e.g., `http://api.com/users/$user_id$`). |
| `header` | Map of HTTP headers. Supports substitution. |
| `body` | The request payload (JSON). Supports substitution in string values. |
| `expected_status` | exact string match for the HTTP status (e.g., `200 OK`, `401 Unauthorized`). |
| `var_to_store` | Map where Key is the *variable name* and Value is the *JSON path* in the response to extract. |

-----

## 🔄 Variable System

### 1\. Usage (Substitution)

You can inject variables into `url`, `header`, and `body` string values by wrapping the variable name in dollar signs: **`$variable_name$`**.

### 2\. Global Variables

Defined at the top level of your JSON file. Useful for tokens or environment settings.

### 3\. Dynamic Extraction (`var_to_store`)

You can extract values from a response body to use in future tests.

  * **Syntax:** Dot notation `parent.child`.
  * **Arrays:** Simple indexing `list[0].id`.
  * **Storage naming:** Variables are stored internally as `test_{testNum}_{variableKey}`. However, the system currently flattens global and stored variables.

*(Note: Ensure your `num` field in the test object matches the execution order for easier debugging).*

-----

## 📝 Example `test.json`

Here is a complete example showing authentication, variable usage, and validation.

```json
{
    "name": "User API Integration Flow",
    "variables": {
        "base_url": "http://localhost:3000",
        "creator_access_token": "eyJhbGciOiJIUzI1Ni..."
    },
    "tests": [
        {
            "num": 1,
            "method": "GET",
            "url": "$base_url$/protected/resource",
            "header": {
                "Authorization": "Bearer $creator_access_token$",
                "Content-Type": "application/json"
            },
            "body": {},
            "expected_status": "200 OK",
            "expected_response": {},
            "var_to_store": {
                "user_id": "data.user.id"
            }
        },
        {
            "num": 2,
            "method": "GET",
            "url": "$base_url$/users/$test_1_user_id$/details",
            "header": {
                "Authorization": "Bearer $creator_access_token$"
            },
            "body": null,
            "expected_status": "200 OK"
        }
    ]
}
```

## ⚠️ Known Limitations

1.  **Strict JSON**: The configuration file must be valid JSON. Trailing commas are not allowed.
2.  **Variable Scope**: Extracted variables are global to the runtime of the suite once stored.

## 🤝 Contributing

1.  Fork the project.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.
