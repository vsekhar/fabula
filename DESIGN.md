# Design thoughts

 * Log (prove something did exist at some point in the past)
 * Mutex (prove something did not exist at some point in the past)
   * Might not be needed: threhold encryption

## Log

* Hosted Merkle trees
    * Can be more than one in the ecosystem
        * Entities can trust different trees
        * Can log to more than one as part of a protocol
            * Timestamp is latest of all timestamps
* Merkle weave
    * Scale writes via cross-linked prefix Merkle trees
    * MMRs for easier seeking etc.
    * Summarize within a prefix tree by bagging peaks
    * Summarize the weave by comining prefix tree summaries
    * Prove entry within a prefix tree
    * Prove entry within the weave

## Expiry

Handling expiry of transactions.

* Initiator logs transaction and desired timeout duration to all logs
* Logs return their timestamps
* Initiator uses latest timestamp from logs as timestamp of transaction
    * Expiry countdown starts
* Initiator sends timestamps to all participants
    * Participants verify that their log was consulted and that the chosen timestamp is at or after the timestamp returned by their trusted log
* Participants log their commitments to their trusted logs, return log entries to initiator
    * All commitments must be logged before the timeout
    * Early logs shorten the timeout for the participants that trust it
* Initiator collects commitments
* If all needed commitments are collected, initator logs completion to all logs
    * All logs must return a timestamp within the timeout window
    * If any log returns a timestamp outside the expiry window, the transaction is considered Expired.

Open questions:

* What happens if initiator vanishes? Or participant vanishes?
* Initiator can fork timeline:
    * Return one set of responses to a few participants (e.g. showing valid completetion) and another set of responses to other participants (e.g. showing expiration)
    * Expiration is an exclusion property: need to prove something _didn't_ happen or need to prove something did

## Threshold crypto

* Initiator logs some random cyphertext, distributes keys to participants
    * Initiator doesn't need to know what cleartext is
    * We only need to later validate that the eventual cleartext (likely random also) is a valid decryption of random cyphertext
* Participants contribute to decryption only of they are voting to complete the transaction
* The existance of a valid cyphertext means all participants at some point decrypted their portion
* Decryption should involve timestamps and be logged
* Expiry should indicate who was the odd one out
* Guarantees:
    * If a valid cleartext exists for the cyphertext anywhere (in the possession of any participant or in any log) then all participants must have voted to commit.
    * If a participant dies or does not contribute to the decryption, it can be assured the cleartext does not exist anywhere. I.e. that no one can claim that the transaction completed.
* Same as signature of vote.
* Real issue is expiry: provable non-action of participants
    * Who decides?
    * E.g. can use threshold crypto to reveal to initiator the cleartext, but then the initiator can die or act like the last participant failed and stoke an expiry.
* See also:
    * Distributed key generation
    * Group digital signatures
    * [God protocols](https://web.archive.org/web/20070927012453/http://www.theiia.org/ITAudit/index.cfm?act=itaudit.archive&fid=216)
