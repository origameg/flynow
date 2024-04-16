# flynow

## Introduction
I love to travel and spend a lot of time researching flight options whenever we are planning a trip. Whenever I watch travel competition shows like _The Amazing Race_ and _JetLag: The Game_, I'm always struck by the idea of booking a flight right before it departs. Surely that must cost a fortune?

**flynow** is an initial attempt to create a utility that could be used in such unusual travel scenarios. Or for someone with the spontanaity to hop on a flight later the same day. Or, more realistically, for someone like me who's more likely to sit at her computer and just imagine doing such things!

## How it works
The application fetches a list of realtime flight information from AviationStack for the given departure airport, filtering on flights in the "Scheduled" state to skip over any that have already departed. From this data, it can determine a list of possible destination airports.

For each of these destinations in parallel, it sends a flight booking search to Amadeus and identifies the cheapest flight from the result. Once the complete set of searches is complete, it displays the results.

## Limitations
There are several restrictions in the free versions of the APIs. Most notably, aviationstack allows only 100 requests per month, so the application can only run 100 times in a month before that limit is reached. And they have my credit card number! 🙈 Because of this limit, I only consume the first page of their results and didn't do much exploration around optimizing the use of that API.

The Amadeus API uses cached data, so the flight prices would need to be reconfirmed before relying on them.

## Future improvements
### Make it a REST API
Originally, I wanted to build this as a REST API with a few of its own query parameters. Unfortunately, with so much to learn, I decided to prioritize the mutlithreading and API consumption, and this didn't make it into the MVP.

### Finish the unit testing
My heart aches to produce code without unit tests, but my learning style didn't align too well with TDD, and the testing had to take a back seat. As I became more comfortable with the language and code structure, I tried to focus more on testability, but there is definitely a lot of room for improvement there.

### Replace the hard-coded secrets
This probably goes without saying! 😆🔐
