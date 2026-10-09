// SPDX-License-Identifier: Apache-2.0
// Part of "Advanced Golang". Code and tests are Apache-2.0; see LICENSE-CODE.

// Package practice is the week 1 practice session: a bounded parallel
// executor, built in six steps and measured across -cpu=1,2,8. See
// run.sh for the commands.
package practice

import (
	sequential "course/w01/practice/step0-sequential"
	unbounded "course/w01/practice/step1-unbounded"
	semaphore "course/w01/practice/step2-semaphore"
	chunked "course/w01/practice/step3-chunked"
	workerpool "course/w01/practice/step4-workerpool"
	prefilled "course/w01/practice/step5-prefilled"
)

// Step is one version of the executor. Every step has the same signature:
// run tasks with at most n in flight, and return when all have finished.
type Step struct {
	Name string
	Run  func(n int, tasks []func())
}

// Steps lists the executors in the order the session builds them.
var Steps = []Step{
	{"0-sequential", sequential.Run},
	{"1-unbounded", unbounded.Run},
	{"2-semaphore", semaphore.Run},
	{"3-chunked", chunked.Run},
	{"4-workerpool", workerpool.Run},
	{"5-prefilled", prefilled.Run},
}
