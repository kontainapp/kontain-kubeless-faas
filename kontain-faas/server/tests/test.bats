#!/usr/bin/env bats

FAAS_ADDR=$(minikube ip)

@test "test_func_data_with_hc" {
  run curl http://${FAAS_ADDR}/test_func_data_with_hc
}
