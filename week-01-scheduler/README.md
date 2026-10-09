# Week 1: Goroutines and the scheduler

How Go runs many goroutines on a few OS threads: G, M and P, run queues, work stealing,
what happens when a goroutine blocks in a syscall, preemption, and where GOMAXPROCS
comes from.

![Three threading models: 1:1, N:1 and M:N](threading-models.svg)

Three ways to map user threads to OS threads. Go picked M:N.

![G, M and P as a picture](gmp.svg)

Two Ps, each with its M, running G, runnext slot and 256-slot local run queue, plus the
global run queue, parked goroutines, idle Ms and sysmon.

The runnable demos are in [demos/](demos/). The practice session, a bounded parallel
executor built in six steps, is in [practice/](practice/). The slides and the lecture
text come later.

## OS background: videos

The lecture assumes some operating systems background and does not teach it: processes and
threads, system calls, virtual memory, signals. If any of that is new or rusty, these free
university lectures cover it.

About 1.5 hours each unless marked. One or two per topic is plenty. Links checked 2026-10-08.

### English

**Processes and threads**
- [Exceptions & Processes](https://www.youtube.com/watch?v=kSErOgVqpAA), CMU 15-213
- [Threads and Processes](https://www.youtube.com/watch?v=QSep-J6vtdc), Berkeley CS162
- [Thread Switching](https://youtu.be/vsgrTHY5tkg), MIT 6.S081

**System calls and epoll**
- [Four Fundamental OS Concepts](https://www.youtube.com/watch?v=4FpG1DcvHzc), Berkeley CS162: user and kernel mode
- [System Call Entry/Exit](https://youtu.be/T26UuauaxWA), MIT 6.S081
- [Concurrent Programming](https://www.youtube.com/watch?v=Wf_zhhI4bNo), CMU 15-213: `select` and `epoll` event loops

**Virtual memory and stacks**
- [Virtual Memory: Concepts](https://www.youtube.com/watch?v=aVoy030KBso), CMU 15-213
- [Page Tables](https://youtu.be/f1Hpjty3TT8), MIT 6.S081

**Signals**
- [Signals](https://www.youtube.com/watch?v=mzg1bi3s-Kk), CMU 15-213

**Bonus**
- [Kernels and High-Level Languages](https://youtu.be/AAtXWGwxI9k), MIT 6.S081: its reading is Biscuit, an OS kernel written in Go

### Russian

**Processes and threads**
- [АКОС-11. Процессы](https://www.youtube.com/watch?v=FW6Gg2lQlLE), MIPT, 1 h
- [АКОС-15. Многопоточность](https://www.youtube.com/watch?v=r4zgz9Cl2I4), MIPT
- [Управление процессами и потоками](https://www.youtube.com/watch?v=QSP_Hb87kLU), CSC

**System calls and epoll**
- [АКОС-7. Прерывания и системные вызовы](https://www.youtube.com/watch?v=bK0OGDR0PAY), MIPT
- [АКОС-16. Неблокирующий ввод-вывод](https://www.youtube.com/watch?v=qJWIE_2UMZ8), MIPT, 1 h
- [Лекция 15 апреля: epoll](https://www.youtube.com/watch?v=TLuWGZts4FI), HSE, 3 h

**Virtual memory and stacks**
- [Управление памятью в операционной системе](https://www.youtube.com/watch?v=wU0SA5D9068), CSC
- [Лекция 11 февраля: mmap](https://www.youtube.com/watch?v=MVhMM0k1EvM), HSE, 2.5 h

**Signals**
- [АКОС-12. Сигналы](https://www.youtube.com/watch?v=Fw-qo4lmnM0), MIPT

**CPU caches** (needed in week 3)
- [АКОС-6. Кеш. Конвейер. SIMD.](https://www.youtube.com/watch?v=L_Cu7Ppx5aQ), MIPT

### Full courses

- CMU 15-213, *Introduction to Computer Systems*, Fall 2017, Randal Bryant and Phil Gibbons ([playlist](https://www.youtube.com/playlist?list=PL2dWYoM7ypKy8yuOV01RGMRDEaw5sNyyz))
- Berkeley CS162, *Operating Systems*, John Kubiatowicz ([playlist](https://www.youtube.com/playlist?list=PLF2K2xZjNEf97A_uBCwEl61sdxWVP7VWC))
- MIT 6.S081, *Operating System Engineering*, Fall 2020 ([schedule with videos](https://pdos.csail.mit.edu/6.S081/2020/schedule.html))
- MIPT, *АКОС*, 2020-21, В. В. Яковлев ([autumn](https://www.youtube.com/playlist?list=PL4_hYwCyhAvZcOr5sJzuLmze2F6wPms-A), [spring](https://www.youtube.com/playlist?list=PL4_hYwCyhAvalZgPVlOhi7QkdQTD-x-PM))
- HSE, *АКОС*, 2022-23, Александр Мыльцев ([playlist](https://www.youtube.com/playlist?list=PLEwK9wdS5g0p7ihPzK8FXf_WZNNdrUbnm)), about 3 h per session
- Computer Science Center, *Введение в архитектуру ЭВМ и элементы ОС*, 2013 ([playlist](https://www.youtube.com/playlist?list=PLlb7e2G7aSpRZ9wDzXI-VYpk59acLFOIr)), older and gentler, a good first pass
