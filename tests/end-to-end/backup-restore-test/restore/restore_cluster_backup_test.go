package restore

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

const functionYaml = "function.yaml"
const functionName = "hello"

var testUUID = uuid.New()
var namespace = "restore-test-" + testUUID.String()
var backupName = "test-" + testUUID.String()

func TestBackupAndRestoreCluster(t *testing.T) {
	Convey("Setup", t, func() {
		Convey("Create namespace", func() {
			So(createNamespace(namespace), ShouldContainSubstring, "created")
			Convey("Create function", func() {
				createFunction(functionYaml, namespace)
				So(func() string {
					return getFunctionState(namespace, functionName)
				}, ShouldReturnSubstringEventually, "Running", 60*time.Second, 1*time.Second, func() string { return printLogsFunctionPodContainers(namespace, functionName) })
				So(func() string {
					host := fmt.Sprintf("http://%s.%s:8080", functionName, namespace)
					value, err := getFunctionOutput(host, namespace, functionName)
					if err != nil {
						t.Fatalf("Could not get function output for host%v: %v", host, err)
					}
					return value
				}, ShouldReturnSubstringEventually, testUUID.String(), 60*time.Second, 1*time.Second)
			})
		})
	})

	Convey("Backup Cluster", t, func() {
		So(func() string {
			stdout, stderr, err := runCommand("ark", []string{"backup", "create", backupName, "--include-namespaces", namespace})
			if err != nil {
				t.Fatalf("Error was: %v \n %v", err, stderr.String())
			}
			return stdout.String()
		}(), ShouldContainSubstring, "submitted successfully")

		Convey("Check backup status", func() {
			So(func() string {
				stdout, stderr, err := runCommand("ark", []string{"backup", "get", backupName, "-oyaml"})
				if err != nil {
					t.Fatalf("Error was: %v \n %v", err, stderr.String())
				}
				return stdout.String()
			}, ShouldReturnSubstringEventually, "phase: Completed", 60*time.Second, 1*time.Second)
			Convey("Delete resources from cluster", func() {
				_, stderr, _ := runCommand("kubectl", []string{"delete", "ns", namespace, "--grace-period=0", "--force", "--ignore-not-found"})
				So(stderr.String(), ShouldNotContainSubstring, "Error from server (Conflict): Operation cannot be fulfilled on namespaces")
				So(stderr.String(), ShouldNotContainSubstring, "No resources found")
				So(func() string {
					_, stderr, _ := runCommand("kubectl", []string{"get", "ns", namespace, "-oyaml"})
					return stderr.String()
				}, ShouldReturnSubstringEventually, "NotFound", 1*time.Minute, 1*time.Second)
			})

		})

	})

	Convey("Restore Cluster", t, func() {
		So(func() string {
			stdout, stderr, err := runCommand("ark", []string{"restore", "create", "--from-backup", backupName, backupName})
			if err != nil {
				t.Fatalf("Error was: %v \n %v", err, stderr.String())
			}
			return stdout.String()
		}(), ShouldContainSubstring, "submitted successfully.")
		So(func() string {
			stdout, _, _ := runCommand("ark", []string{"restore", "get", backupName})
			return stdout.String()
		}, ShouldReturnSubstringEventually, "Completed", 5*time.Minute, 1*time.Second)
		Convey("Test restored resources", func() {
			So(func() string {
				return getFunctionState(namespace, functionName)
			}, ShouldReturnSubstringEventually, "Running", 60*time.Second, 1*time.Second, func() string { return printLogsFunctionPodContainers(namespace, functionName) })
			So(func() string {
				value, err := getFunctionOutput(fmt.Sprintf("http://%s.%s:8080", functionName, namespace), namespace, functionName)
				if err != nil {
					t.Fatalf("Could not get function output: %v", err)
				}
				return value
			}, ShouldReturnSubstringEventually, testUUID.String(), 60*time.Second, 1*time.Second)
		})
	})
}
