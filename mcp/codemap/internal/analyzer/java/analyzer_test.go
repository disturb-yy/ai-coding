package java

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/disturb-yy/codemap/internal/model"
)

func TestAnalyzeSpringProject(t *testing.T) {
	root := t.TempDir()
	writeJavaFile(t, root, "pom.xml", `<project></project>`)
	writeJavaFile(t, root, "src/main/java/com/example/order/OrderController.java", `package com.example.order;

import com.example.payment.PaymentService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/orders")
public class OrderController {
    private final PaymentService paymentService;

    public OrderController(PaymentService paymentService) {
        this.paymentService = paymentService;
    }

    @GetMapping("/{id}")
    public OrderDto getOrder(String id) {
        validate(id);
        return paymentService.load(id);
    }

    @PostMapping
    public OrderDto create(OrderDto dto) {
        return paymentService.charge(dto);
    }

    private void validate(String id) {}
}
`)
	writeJavaFile(t, root, "src/main/java/com/example/order/OrderDto.java", `package com.example.order;

public class OrderDto {}
`)
	writeJavaFile(t, root, "src/main/java/com/example/payment/PaymentService.java", `package com.example.payment;

public interface PaymentService {
    PaymentReceipt load(String id);
    PaymentReceipt charge(Object request);
}
`)
	writeJavaFile(t, root, "src/main/java/com/example/payment/PaymentReceipt.java", `package com.example.payment;

public class PaymentReceipt {}
`)

	project, err := New().Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	order := findModule(project.Modules, "com/example/order")
	if order == nil {
		t.Fatalf("missing order module: %#v", project.Modules)
	}
	if !contains(order.Dependencies, "com/example/payment") {
		t.Fatalf("order deps = %v, want com/example/payment", order.Dependencies)
	}
	if !contains(order.ExportedTypes, "OrderController") || !contains(order.ExportedTypes, "OrderDto") {
		t.Fatalf("order exported types = %v", order.ExportedTypes)
	}
	if !contains(order.ExportedMethods, "OrderController.getOrder") || !contains(order.ExportedMethods, "OrderController.create") {
		t.Fatalf("order exported methods = %v", order.ExportedMethods)
	}

	payment := findModule(project.Modules, "com/example/payment")
	if payment == nil {
		t.Fatal("missing payment module")
	}
	if !contains(payment.KeyInterfaces, "PaymentService") {
		t.Fatalf("payment interfaces = %v, want PaymentService", payment.KeyInterfaces)
	}

	if !hasRoute(project.Routes, "GET", "/orders/{id}", "com/example/order") {
		t.Fatalf("missing GET /orders/{id}; routes=%#v", project.Routes)
	}
	if !hasRoute(project.Routes, "POST", "/orders", "com/example/order") {
		t.Fatalf("missing POST /orders; routes=%#v", project.Routes)
	}
	if !hasCallEdge(project.CallEdges, "com/example/order", "OrderController.getOrder", "com/example/payment", "paymentService.load") {
		t.Fatalf("missing paymentService.load edge; edges=%#v", project.CallEdges)
	}
	if !hasCallEdge(project.CallEdges, "com/example/order", "OrderController.getOrder", "com/example/order", "OrderController.validate") {
		t.Fatalf("missing local validate edge; edges=%#v", project.CallEdges)
	}
	if len(project.Flows) != 1 || project.Flows[0].Trigger != "com/example/order" {
		t.Fatalf("flows=%#v, want one flow from order to payment", project.Flows)
	}
}

func TestAnalyzeJavaProjectDetectsEmbeddedPersistence(t *testing.T) {
	root := t.TempDir()
	writeJavaFile(t, root, "src/main/java/com/example/order/OrderService.java", `package com.example.order;

import java.sql.Connection;
import java.sql.PreparedStatement;
import java.sql.SQLException;

public class OrderService {
    public void save(Connection connection, String id) throws SQLException {
        PreparedStatement statement = connection.prepareStatement("insert into orders(id) values (?)");
        statement.setString(1, id);
        statement.executeUpdate();
    }
}
`)

	project, err := New().Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	order := findModule(project.Modules, "com/example/order")
	if order == nil {
		t.Fatalf("missing order module: %#v", project.Modules)
	}
	if !contains(order.Dependencies, "storage/database") {
		t.Fatalf("order deps = %v, want storage/database", order.Dependencies)
	}

	storage := findModule(project.Modules, "storage/database")
	if storage == nil {
		t.Fatalf("missing virtual storage module: %#v", project.Modules)
	}
	if storage.Name != "data" {
		t.Fatalf("storage module name = %q, want data", storage.Name)
	}
}

func TestAnalyzeSpringProjectWithMultilineAnnotations(t *testing.T) {
	root := t.TempDir()
	writeJavaFile(t, root, "src/main/java/com/example/order/OrderController.java", `package com.example.order;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping(
    path = "/api/v1/orders"
)
public class OrderController {
    @GetMapping(
        path = "/{id}"
    )
    public OrderDto getOrder(String id) {
        return new OrderDto();
    }
}
`)
	writeJavaFile(t, root, "src/main/java/com/example/order/OrderDto.java", `package com.example.order;

public class OrderDto {}
`)

	project, err := New().Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if !hasRoute(project.Routes, "GET", "/api/v1/orders/{id}", "com/example/order") {
		t.Fatalf("missing multiline annotation route; routes=%#v", project.Routes)
	}
}

func TestAnalyzeSpringProjectWithMultipleAnnotationPaths(t *testing.T) {
	root := t.TempDir()
	writeJavaFile(t, root, "src/main/java/com/example/user/UserController.java", `package com.example.user;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/users")
public class UserController {
    @GetMapping(path = {"/active", "/enabled"})
    public UserDto listActive() {
        return new UserDto();
    }
}
`)
	writeJavaFile(t, root, "src/main/java/com/example/user/UserDto.java", `package com.example.user;

public class UserDto {}
`)

	project, err := New().Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if !hasRoute(project.Routes, "GET", "/users/active", "com/example/user") {
		t.Fatalf("missing /users/active route; routes=%#v", project.Routes)
	}
	if !hasRoute(project.Routes, "GET", "/users/enabled", "com/example/user") {
		t.Fatalf("missing /users/enabled route; routes=%#v", project.Routes)
	}
}

func writeJavaFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func findModule(mods []*model.Module, path string) *model.Module {
	for _, mod := range mods {
		if mod.Path == path {
			return mod
		}
	}
	return nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func hasRoute(routes []*model.Route, method, path, module string) bool {
	for _, route := range routes {
		if route.Method == method && route.Path == path && route.Module == module {
			return true
		}
	}
	return false
}

func hasCallEdge(edges []*model.CallEdge, callerModule, callerFunc, calleeModule, calleeFunc string) bool {
	for _, edge := range edges {
		if edge.CallerModule == callerModule &&
			edge.CallerFunc == callerFunc &&
			edge.CalleeModule == calleeModule &&
			edge.CalleeFunc == calleeFunc {
			return true
		}
	}
	return false
}
