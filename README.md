# Target Audience

In not necessarily prioritized order:
* Turbine engineers
* Turbine dev rel/implementation engineres
* Turbine customers

# Use Cases

In roughly prioritized order:

1. CRUD operations on core objects (zones, clusters, domains, shared rules, routes, proxies)
2. Higher level operations
  1. Initial environment setup
  2. Execute the test workflow
  3. Execute the release workflow
3. Generate nginx config for a proxy
4. Pull stats for a given object

# General Approach

## CRUD operations

Not to totally ape kubernetes, but I’ve been pretty happy with kubectl. Supporting the following kubectl commands makes sense:

* get - gets a list of resources, e.g. `tbnctl get zones`. Equivalent to `GET <api>/v1.0/<object type>`
* describe - gets details on a single resource, e.g. `tbnctl describe zone testbed`. Equivalent to `GET <api>/v1.0/<object type>/<object id>`
* create - create a new resource from a file/flags, e.g. `tbnctl create zone testbed -f testbed.json`. Equivalent to `POST <api>/v1.0/<object type>`
* edit - modify a resource, e.g. `tbnctl edit zone testbed`. Equivalent to `PUT <api>/v1.0/<object type>/<object id>`
* delete - deletes a single resoucre, e.g. `tbnctl delete zone testbed`. Equvalent to `DELETE <api>/v1.0/<object type>/<object id>`

Most of these are pretty straightforward. Create and edit get a bit trickier since we need to provide more detailed JSON payloads. For create I think we just take a file as input. Extra credit for allowing all args to be specified on the command line. For edit, the kubectl trick of writing out a tmp file, popping open the editor and then posting back the edited content seems pretty safe though.

In addition to these commands it would be cool to support describe, which spits out doc on a given object, e.g. tbnctl describe zone. I think we can generate all this from swagger. I think.

## Initial environment setup

Even with CRUD operations supported there is still a somewhat tedious list of objects that need to be created to have a minimal functioning environment. Adding `tbnctl init <zone name>` should create the zone, a domain, a default cluster, a reasonable default shared rules, a proxy and a route.

Executing the test and release workflows is likely a later step, although it might be useful to explore the UX via the CLI before we put a ton of effort into visual design.

## Nginx config generation

For use case 2, generating an nginx config should also be pretty straightforward, something like `tbnctl getconfig <proxy>`.

## Stats

Stats is probably trickier, and lower priority so probably doesn’t make a ton of sense to design here.
