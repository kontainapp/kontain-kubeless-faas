#!/usr/bin/env bats

FAAS_ADDR=$(minikube ip)

@test "test_func_data_with_hc" {
  run curl -s http://${FAAS_ADDR}/test_func_data_with_hc
  [ $status -eq 0 ]
  echo "$output"
  [ "$output" == "the quick brown fox" ]
}
