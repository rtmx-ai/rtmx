# REQ-LANG-008: Java/JVM Testing Integration

## Metadata
- **Category**: LANG
- **Subcategory**: Java
- **Priority**: MEDIUM
- **Phase**: 18
- **Status**: MISSING
- **Dependencies**: REQ-LANG-007

## Requirement

RTMX shall provide Java testing integration via JUnit 5 extensions supporting requirement annotations on test methods.

## Rationale

Java remains dominant in enterprise software development where requirements traceability is often mandated by compliance frameworks (SOX, HIPAA, etc.).

## Design

### Installation

```xml
<!-- Maven -->
<dependency>
    <groupId>ai.rtmx</groupId>
    <artifactId>rtmx-junit5</artifactId>
    <version>0.1.0</version>
    <scope>test</scope>
</dependency>
```

```groovy
// Gradle
testImplementation 'ai.rtmx:rtmx-junit5:0.1.0'
```

### Annotation Syntax

```java
import ai.rtmx.Req;
import ai.rtmx.RtmxExtension;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(RtmxExtension.class)
class AuthenticationTest {

    @Test
    @Req("REQ-AUTH-001")
    void testLoginSuccess() {
        // test implementation
    }

    @Test
    @Req(value = "REQ-AUTH-002", scope = "integration", technique = "boundary")
    void testLoginInvalidPassword() {
        // test implementation
    }

    // Multiple requirements
    @Test
    @Req("REQ-AUTH-001")
    @Req("REQ-AUDIT-001")
    void testLoginAudited() {
        // test implementation
    }
}
```

### Kotlin Support

```kotlin
import ai.rtmx.Req
import org.junit.jupiter.api.Test

class AuthenticationTest {

    @Test
    @Req("REQ-AUTH-001")
    fun `login should succeed with valid credentials`() {
        // test implementation
    }
}
```

### Output Integration

```bash
# Maven
mvn test -Drtmx.output=rtmx-results.json

# Gradle
./gradlew test -Prtmx.output=rtmx-results.json

# Or via rtmx
rtmx verify --command "mvn test"
```

### JUnit 4 Support

```java
import ai.rtmx.Req;
import ai.rtmx.RtmxRule;
import org.junit.Rule;
import org.junit.Test;

public class AuthenticationTest {

    @Rule
    public RtmxRule rtmx = new RtmxRule();

    @Test
    @Req("REQ-AUTH-001")
    public void testLoginSuccess() {
        // test implementation
    }
}
```

## Acceptance Criteria

1. Package available on Maven Central
2. JUnit 5 extension works with `@Req` annotation
3. JUnit 4 Rule available for legacy projects
4. Kotlin tests work seamlessly
5. Test results output compatible JSON format
6. `rtmx verify --command "mvn test"` correctly updates status

## Test Strategy

- Unit tests for annotation processing
- Integration tests with JUnit 5 and JUnit 4
- Kotlin interop tests
- Maven and Gradle build integration tests

## Package Structure

```
rtmx-junit5/
├── pom.xml
├── src/
│   ├── main/java/ai/rtmx/
│   │   ├── Req.java              # @Req annotation
│   │   ├── RtmxExtension.java    # JUnit 5 extension
│   │   └── RtmxRule.java         # JUnit 4 rule
│   └── test/java/ai/rtmx/
│       └── RtmxExtensionTest.java
└── examples/
    ├── junit5/
    └── junit4/
```

## References

- JUnit 5 Extension Model
- JUnit 4 Rules
- Maven Central publishing
- REQ-LANG-007 marker specification
