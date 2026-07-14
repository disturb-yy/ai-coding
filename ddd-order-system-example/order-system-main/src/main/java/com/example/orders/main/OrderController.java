package com.example.orders.main;

import com.example.orders.api.CreateOrderCommand;
import com.example.orders.api.OrderCommandService;
import com.example.orders.api.OrderView;
import java.math.BigDecimal;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/orders")
public class OrderController {

    private final OrderCommandService orderCommandService;

    public OrderController(OrderCommandService orderCommandService) {
        this.orderCommandService = orderCommandService;
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public OrderView create(@RequestBody CreateOrderRequest request) {
        return orderCommandService.create(new CreateOrderCommand(request.customerId(), request.totalAmount()));
    }

    @PostMapping("/{orderId}/confirm")
    public OrderView confirm(@PathVariable String orderId) {
        return orderCommandService.confirm(orderId);
    }

    @GetMapping("/{orderId}")
    public OrderView get(@PathVariable String orderId) {
        return orderCommandService.get(orderId);
    }

    public record CreateOrderRequest(String customerId, BigDecimal totalAmount) {
    }
}
