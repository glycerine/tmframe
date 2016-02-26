# TMFRAME

TMFRAME, pronounced "time frame", is a simple and efficient
binary standard for encoding time series data.

Starting with a 64-bit nanoseconds-since the Unix epoch
timestamp, the idea here is that the low 3-bits are really
just random noise, given that our clock calibrations just
aren't that accurate.

So we replace those 3 bits with a useful data
payload to get a highly compressed timeseries format.

VERSION: This spec is version 2016-Feb-26, which
revises the original definition of PTI=1 by changing PTI=1
into a 16 byte message to support a single int64 payload.
This change facilitates indexing. The convention for
boolean series is now to use PTI=4 for true and PTI=0
for false.

specification
=============

The TMFRAME format allows very compact expression of time-series.
For example, for a simple time-series, the TMFRAME encoding
can be as simple as a sequence of 64-bit timestamps (whose
resolution is limited to 10 nanoseconds). However the same
format can be accompanied by much longer additional
event data if need be. Common situations where a single
float64 are needed for the timepoint's value are supported
with exactly two words (two 64-bit words; one for the
timestamp and one for the float64 payload).

# overview of the format

A TMFRAME message always starts with a primary word.

Depending on the content of the low 3 bits of the primary word,
the primary word may be the entire message.
However, there may also be additional words and bytes following the
primary word that complete the message.

TMFRAME messages can be classified as being either be 8 bytes
(primary word only), 16 bytes long, greater than 16 bytes long.

Frequently a TMFRAME message will consist of one primary word,
one UDE word, and a variable length payload.

The primary word and UDE word are always 64-bit words each. The payload
can be up to 2^43 bytes in length.

We illustrate the possible TMFRAME message lengths here:

a) primary word only

~~~
+---------------------------------------------------------------+
|      primary word (64-bits) with PTI={0, 4, 5, or 6}       |
+---------------------------------------------------------------+
~~~

b) primary word and UDE word only:

~~~
+---------------------------------------------------------------+
|                primary word (64-bits) with PTI=7              |
+---------------------------------------------------------------+
|            User-defined-encoding (UDE) descriptor             |
+---------------------------------------------------------------+
~~~

c) primary word + UDE word + variable byte-length message:

~~~
+---------------------------------------------------------------+
|                primary word (64-bits) with PTI=7              |
+---------------------------------------------------------------+
|            User-defined-encoding (UDE) descriptor             |
+---------------------------------------------------------------+
|               variable length                                 |
|                message here                          ----------
|     (the UDE supplies the exact byte-count)          |
+-------------------------------------------------------
~~~

There are also two special payload types that are not UDE based.
They handle the common case of attaching one or two
64-bit values to a timestamp.

d) primary word + one int64

~~~
+---------------------------------------------------------------+
|                primary word (64-bits) with PTI=1              |
+---------------------------------------------------------------+
|                     V1 (int64)                                |
+---------------------------------------------------------------+
~~~


e) primary word + one float64

~~~
+---------------------------------------------------------------+
|                primary word (64-bits) with PTI=2              |
+---------------------------------------------------------------+
|                     V0 (float64)                              |
+---------------------------------------------------------------+
~~~

f) primary word + one float64 + one int64

~~~
+---------------------------------------------------------------+
|                primary word (64-bits) with PTI=3              |
+---------------------------------------------------------------+
|                     V0 (float64)                              |
+---------------------------------------------------------------+
|                     V1 (int64)                                |
+---------------------------------------------------------------+
~~~


# 1. number encoding rules

Integers and floating point numbers are used in the
protocol that follows, so we fix our definitions of these.

 * Integers: are encoded in little-endian format. Signed integers
    use two’s complement. Integers are signed unless otherwise
    noted.
 * float64, also known as 64-bit floating-point numbers: Encoded
   in little-endian IEEE-754 format.

# 2. primary word encoding

A TMFRAME message always starts with a primary word.

~~~

msb                  primary word (64-bits)                   lsb
+-----------------------------------------------------------+---+
|                        TMSTAMP                            |PTI|
+-----------------------------------------------------------+---+

TMSTAMP (61 bits) =
     The primary word is generated by starting
     with a 64-bit signed little endian integer, the number
     of nanoseconds since the unix epoch; then truncating off
     the lowest 3-bits and overwriting them with the value of PTI.
     The resulting TMSTAMP value is the 61 most significant
     bits of the timestamp and can be used directly as an
     integer timestamp by first copying the full 64-bits of the
     timeframe word and then zero-ing out the 3 bits of PTI.
     
PTI (3 bits) = Payload type indicator, decoded as follows:

    0 => a zero value is indicated for this timestamp.
         (the zero value can also be encoded, albeit
         less efficiently, by a UDE word with bits all 0).
         
         Use the zero-value for time-stamp only time-series.

         The primary word is the only word in this message.
         The next word will be the primary word of the next
         message on the wire.

         By convention, the 0 value can indicate the
         payload false for boolean series.

    1 => exactly one 64-bit int64 payload value follows.
         The message has exactly two 64-bit words.
         The payload is known as V1.

    2 => exactly one 64-bit float64 payload value follows.
         Nmemonic: The total number of 64-bit words in the message is 2.
         The payload is known as V0.

    3 => exactly two 64-bit payload values follow, one float64 and one int64.
         Nmemonic: The total number of 64-bit words in the message is 3.
         The payload components are known as V0 (the float64), and
         V1 (the int64).

    4 => NULL: the null-value, a known and intentionally null value. Written as NULL.

         NB By convention, for a strictly boolean series, PTI=4 is the true value,
         while PTI=0 is the false value.

         The primary word is the only word in this message.

    5 => NA: not-available, an unintentionally missing value.
         In statistics, this indicates that *any* value could
         have been the correct payload here, but that the
         observation was not recorded. a.k.a. "Missing data". Written as NA.

         The primary word is the only word in this message.

    6 => NaN: not-a-number, IEEE-754 floating point NaN value.
         Obtained when dividing zero by zero, for example. math.IsNaN()
         detects these.

         The primary word is the only word in this message.

    7 => user-defined-encoding (UDE) descriptor word follows.

~~~

# 3. User-defined-encoding descriptor

~~~
msb    user-defined-encoding (UDE) descriptor 64-bit word     lsb
+---------------------------------------------------------------+
| EVTNUM (21-bits)  |                UCOUNT (43-bits)           |
+---------------------------------------------------------------+

  UCOUNT => is a 43-bit unsigned integer number of bytes that
       follow as a part of this message. Zero is allowed as a
       value in UCOUNT, and is useful when the type information in EVTNUM
       suffices to convey the event. Mask off the high 21-bits
       of the UDE to erase the EVTNUM before using the count
       of bytes found in UCOUNT. The payload starts immediately
       after the UDE word, and can be up to 8TB long (2^43 bytes).
       Shorter payloads are recommended whenever possible.

       There is no requirement that UCOUNT be padded to
       any alignment boundary. It should be the exact length
       of the payload in bytes.

       The next message's primary word will commence after the
       UCOUNT bytes that follow the UDE.

       If UCOUNT is > 0, then the payload of bytes must
       include a 0 byte as its last value. This assists
       in languages bindings (e.g. C) where strings need a
       terminating zero byte.

  EVTNUM => a 21-bit signed two's-compliment integer capable
       of expressing values in the range [-(2^20), (2^20)-1].

       Positive numbers are for pre-defined system event
       types. Negative numbers are reserved for user-defined
       event types starting with -2, -3, -4, ...

       There is one pre-defined user-defined event number.
       The one pre-defined user EVTNUM value is:

       -1 => an error message string in utf8 follows; it is
             of length UCOUNT, and the count includes a
             zero termination byte if and only if the string has
             one or more bytes in it.

       Any custom user-defined types added by the user will
       therefore start at EVTNUM = -2. The last usable EVTNUM is
       the -1 * (2^20) value; so over one million user
       defined event types are available.

       System defined EVTNUM values as of this writing are:

       0 => this is also a zero value payload. The corresponding
            UCOUNT must also be 0. There are no other words
            in this message. This allows encoders to not
            have to go back and compress out a zero value by
            writing a PTI of zero; although they are encouraged
            to do so whenever possible to save a word of space.

       1..7 => these are reserved and should never actually
               appear in the EVTNUM field on the wire. However
               we defined them here to make the API
               easier to use--when NewFrame() is called
               with evtnum in this range, the PTI field is automatically
               set accordingly, and the implicit compression
               of the primary word-only messages is recognized.
               To provide documentation for the NewFrame() calls
               evtnum formal parameter, we will restate the encoding here:

               1 => EvOneInt64, the payload is defined to be the
                    following int64 value.
               2 => EvOneFloat64, the payload will be the
                    following float64 value.
               3 => EvTwo64, the payload will be the float64
                    and the int64 that follow.
               4 => EvNull, payload is defined as the NULL value.
               5 => EvNA, payload is defined as NA, or Not-Available,
                    denoting missing data.
               6 => EvNaN, payload denotes IEEE-754 Not-a-Number, NaN.
               7 => EvUDE, payload is describe by the UDE word that
                    follows.

       8 => a TMFRAME-HEADER value follows, giving time-series
            metadata. To be described later.
            
       9 => a Msgpack encoded message follows.
       
       10 => a Binc encoded message follows.
       
       11 => a Capnproto encoded message segment follows.

       12 => a sequence of S-expressions (code or data) in zygomys
            parse format follows. [note 1]
 
       13 => the payload is a UTF-8 encoded string. As noted
             above in the UCOUNT section, the wire
             format will include a zero termination byte after
             the string to help with C bindings, and
             the UCOUNT will reflect that inclusion if the
             string length itself is greater than zero.

       14 => the payload is a JSON UTF-8 string. The UCOUNT
             will include an additional terminating zero byte
             if the string has length > 0.
~~~

After any variable length payload that follows the UDE word, the
next TMFRAME message will commence with its primary word.

This concludes the specification of the TMFRAME format.

# conclusion

TMFRAME is a very simple yet flexible format for time series data. It allows
very compact and dense information capture, while providing the
ability to convey and attach full event information to each timepoint as
required.

### implementation

There is a full Go implementation in this repo. [Docs here](https://godoc.org/github.com/glycerine/tmframe).

### notes

[1] For zygomys parse format, see [https://github.com/glycerine/zygomys](https://github.com/glycerine/zygomys)


Copyright (c) 2016, Jason E. Aten.

LICENSE: MIT license.
