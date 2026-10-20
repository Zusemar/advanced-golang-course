# Week 2: Channels and select

How a channel is built (`hchan`, the ring buffer, the wait queues), what a direct handoff
between goroutines costs, how `select` polls and locks its cases, the rules for closed and
nil channels, and what the race detector does and does not see.

The runnable demos are in [demos/](demos/). The practice session, six steps from fan-in to
drain-on-close, is in [practice/](practice/). The homework, `Broadcaster`, is in
[homework/](homework/), due Wednesday 28 October, 23:59 MSK. Corrections to the lecture are
in [lecture/corrections/](lecture/corrections/), corrections to the practice in
[practice/corrections/](practice/corrections/). The slides come later.
