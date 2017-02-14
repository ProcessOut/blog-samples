---
title: Automatons for integrity
date: 2017-02-14 14:18 UTC
tags: development testing technical
category: technical
author: Guillaume "gee-m" Merindol
duration: 10 minutes
subcover: /images/blog/payment-flow.png
---

<span style="margin-left: 3em"/> Bugs are always messy. Though they become more of a problem when you’re dealing with anything of value. And what’s more valuable than money? Handling payments is really delicate, one anomaly can have quite the impact: from the simple “wrong amount” bug to the complex exploit. Running tests and in huge quantities is a must, but so is being intelligent about it. How do we do this? Automatons, also known as finite state machines. FSMs can be used to detect anomalies, and thus preserve the integrity of any state-driven system. Here, it's less about using FSMs to solve your problem, but more about using them to **check** on you.

SPLIT_SUMMARY_BEFORE_THIS

---

#### <span style="text-align: left;">Aparté: Quick review of Finite State Machines</span>

Let's go through it word-by-word, but starting with the `M`:

**Machine**: A machine here is the process that takes input (data), treats it, and outputs it. For example, imagine the data here is a transaction, it comes in the machine, the machine eventually sets the status as `captured` and outputs it.

**State**: State is pretty intuitive, it's the different 'statuses' data can be in. For example, the data of a transaction can be in the states `pending`, `authorized`, and `captured` for a transaction. The different states can also be a combination of elements of a transaction, you could define the states of a transaction as: 1. `pending` with `amount=$0`, 2. `pending` with `amount>$0`, 3. `authorized` with `amount>$0`, 4. `captured` with `amount>$0`. So our *machine* takes this data from **state** to **state**.

**Finite**: This just means that our machine, our state-machine, can only be in **one** of a *finite* number of states at the same time. Simple, a transaction can only be `pending`, `authorized`, or `captured` and never in-between or two at the same time. So it's a **finite-state-machine**.

By now you should see it in your head. It should look like this:

![simple-automaton.png](https://raw.githubusercontent.com/ProcessOut/blog-samples/master/automatons/simple-automaton.png "simple automaton")

> **Arrow**: transition, **Double circle**: (accept) state,
> **PDG**: pending, **ATH**: authorized, **CPT**: captured

<span style="margin-left: 3em"/> So you can see in the drawing you start of at the initial state, **PDG**, from there you can only go to **ATH**, and from there you, wait you can go to both **ATH** and **CPT**? Well yeah, if the authorization expires it'd be better if the customers could try the authorization again. So, that's why there's a redundant arrow. Something else here is that all the states are *accept* states. It means the data is in a *valid* state if it stops in any of the accept state. Some automatons have states where the data can't end on. For example here I could make the **ATH** state not be an accept state, so that the transactions would have to be always captured, because you can't stop on **ATH**, but only on the *final* state **CPT**.

<span style="margin-left: 3em"/> In short, a FSM is defined by a list of its states which include its initial and final state and the conditions for each transition. It's for handling data in an assured and simplified manner. But, it can also be used to check the evolution of your data to make sure the integrity of a flow is preserved, and that no bugs happen.

---

<span style="margin-left: 3em"/>The entire power of FSMs is that they offer bug🐛-free checking. This is huge —and in a world where bugs are bound to happen— this almost feels surreal. The only drawback is that you need a *flow* to apply this on. Basically, where your data goes from state to state, which should be pretty much everywhere. You get to describe exactly how your data shifts. Let me give you an example.

#### Example


<span style="margin-left: 3em"/>To keep it simple, let’s say you’re handling payments 😉. The first step is to design your FSM to handle your flow of data. You have a transaction object, which contains generic information: the item, the amount, the currency, maybe the fees, and the status of the transaction. Hmm, alright. How can that be viewed as state-driven data. Simple! The `status`! To keep it simple, we’ll say that the `status` can go from `pending` to `authorized` to `captured`.

Similar to the apparté, we have this simple design in mind:

![simpler-automaton.png](https://raw.githubusercontent.com/ProcessOut/blog-samples/master/automatons/simpler-automaton.png "simpler automaton")

> **Status**: **PDG**=pending, **ATH**=authorized, **CPT**=captured


<span style="margin-left: 3em"/>You can see it now. This is the way this simplified payment system is going to work. Note that once you have defined the transitions, this means it has to happen this way, and no other way. You can always come back and adjust though.

<span style="margin-left: 3em"/>We have the transitions that are `pending` -> `authorized` -> `captured`. But those transitions can also have ‘rules’. For example, you can say that you cannot go from state `pending` to state `authorized` if there’s no delivery address, or dumber, if the amount has changed in between both states. This preserves integrity. Of course you can also have starting rules for the `pending` state (e.g. the amount must be strictly greater than 0). These transition rules combined with the transitions themselves will ensure that no bugs happen, that data integrity failures are possible.

<span style="margin-left: 3em"/>The next step is coding that FSM. It’s the easiest step, trust me. There are plenty of reputable libraries in all the languages that allow you to implement your FSM easily. You can even take a look at the libraries’ code, it’s not very complicated. I recommend that you pick a library which allows you to set rules for the transitions. Often times there are no difference between rules and transitions, since a rule is just a transition condition. Since I’m a Gopher we’ll be looking at an example in Go of integrating our little FSM. I’ve forked a library (https://github.com/ProcessOut/fsm) in order to make this process simpler.

<pre class="styled-code rounded shadowed large-pre"><code class="go">// It is recommended to read the comments ;).

// Transaction which will be the data that flows.
// Has to implement IDer (see `ID()`) for the fsm package.
type Transaction struct {
    Status      string
    Amount      float64
    AdressLine1 string
    // pending, failed bool
}

// ID will return the status as to define the different 
// states. You don't have to return a string here, you 
// can return anything. You could also return a structure
// which would mean that your states would depend on these
// two fields (see states below)
func (t Transaction) ID() fsm.ID { return t.Status }

var (
  // Here we describe the states that are possible
  // It's your data structure with IDable fields defined
  // in `ID()` set to their values.
  statePdg = Transaction{Status: "pending"}
  stateAth = Transaction{Status: "authorized"}
  stateCpt = Transaction{Status: "captured"}
  // Important:
  // If our `ID()` returned the transaction structure 
  // with three fields (Status, pending, failed),
  // we could define states as:
/*
* (if field not defined it defaults to false)
* stateAthP = Transaction{Status: "authorization", pending: true}
* stateAthF = Transaction{Status: "authorization", failed: true}
* stateAthS = Transaction{Status: "authorization", failed: false}
* stateCapF = Transaction{Status: "capture", failed: true}
* stateCapS = Transaction{Status: "capture", failed: false}
*/
   // This allows you to define flows without changing your
   // data structures. Here the flow could be: (see image)
)
</code></pre>

![struct-automaton.png](https://raw.githubusercontent.com/ProcessOut/blog-samples/master/automatons/struct-automaton.png "multiple fields automaton")

> Where **A** = authorization, **C** = capture, **(p)** = pending, **(s)** = success, **(f)** = failed

<span style="margin-left: 3em"/> The transitions are quite simple here, but as in the commented example we could also define more complex transitions that come from more than one data point (and not just the `status`). Now it's time for the transitions and the  transition rules. Note that in a way, a transition from state to state itself is a transition rule, and this is reflected in the package. Functions that define the rules here are called `Guards`.

<pre class="styled-code rounded shadowed large-pre"><code class="go">// machine is our actual FSM
var machine = fsm.Machine{}

// sampleGuard will be a guard for checking if a transition
// can go through
func sampleGuard(start fsm.State, goal fsm.State) error {
    // I() returns the interface (of the transaction)
    if start.I().(Transaction).Amount <= 0 {
        return errors.New("Can't transition, amount is <= 0")
    }
    if start.I().(Transaction).Amount != 
        goal.I().(Transaction).Amount {
        return errors.New("Can't transition, amount is different")
    }
    return nil
}

func init() {
    // Define our machine's rules
    rules := fsm.Ruleset{}

    // Transition: Pending -> Authorized with rule sampleGuard
    rules.AddRule(
        fsm.NewTransition(statePdg, stateAth),
        sampleGuard)
    // Transition: Authorized -> Captured with no additional rules
    rules.AddTransition(
        fsm.NewTransition(stateAth, stateCpt))

    machine.Rules = &rules
}

// testFlow tests the machine and with its rules defined. 
// Serves as an example on how to use the `fsm' package
func testFlow() {
    // Imagine you have your data that's trying to go from
    // pending to authorized.
    pending := Transaction{Status: "pending"}
    authorized := Transaction{Status: "authorized"}

    // Will fail because of the "sampleGuard"
    machine.State = fsm.NewState(pending)
    if err := machine.Transition(fsm.NewState(authorized)); err != nil {
        fmt.Println("1. Transition failed:", err)
    }

    pending.Amount = 1.99
    authorized.Amount = 1.99
    machine.State = fsm.NewState(pending)
    // Will go through because now there's a price
    if err := machine.Transition(fsm.NewState(authorized)); err != nil {
        fmt.Println("2. Transition failed:", err)
    }

    // Now from authorized to authorized (not possible in 
    // current schema)
    // Will fail
    if err := machine.Transition(fsm.NewState(authorized)); err != nil {
        fmt.Println("3. Transition failed:", err)
    }

    // Note that the machine will, of course, not transition
    // if there's an error in the transition.
}
</code></pre>

Output:

<pre class="styled-code rounded shadowed large-pre"><code class="nohighlight">1. Transition failed: Guard failed from pending to authorized: Can't transition, amount is <= 0
3. Transition failed: No rules found for authorized to authorized
</code></pre>


<span style="margin-left: 3em"/> So not only are FSMs really simple to integrate, they’re also really fast. It’s quite important when you have a lot of tests, and don’t wanna wait a while after every change. We applied FSMs here to the logic of transactions, but the sky’s the limit, and I’m really interested as to where you guys find you can apply this type of logic, do not hesitate to speak to me about this: guillaume@processout.com . Sure, you might have to modify your data structures a bit, but you’ll be reaping the benefits of “forced integrity”, a.k.a. no bugs possible.

<span style="margin-left: 3em"/> The advantage is not only in the way that FSMs are coded, it’s also about the way your data moves. If you think about your data movements as if it was an FSM, it’s usually a much cleaner design, and as a result can even improve your coding speed.

### Nice links

- [Package of the code shown here](github.com/ProcessOut/blog-samples/blob/master/automatons/main.go)

- This designer is what I use to draw beforehand: [Online FSM designer](http://madebyevan.com/fsm/)

- The good ol' [Wikipedia link](https://en.wikipedia.org/wiki/Finite-state_machine) on FSMs

- The good ol' [Wikipedia link](https://en.wikipedia.org/wiki/State_pattern) on state machines applied as a design pattern.

- This is about FSMs in general, you need not read this: [Introductory semi-complex paper](http://www.cse.chalmers.se/~coquand/AUTOMATA/book.pdf)

- And [guillaume@processout.com](mailto:guillaume@processout.com), I'd be happy to talk about this article with you.
