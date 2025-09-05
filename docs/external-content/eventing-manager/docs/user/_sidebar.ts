export default [
  { text: 'Module Lifecycle', link: './01-manager' },
  { text: 'Configuration', link: './02-configuration' },
  { text: 'Eventing Architecture', link: './evnt-architecture' },
  { text: 'Event Names', link: './evnt-event-names' },
  { text: 'Eventing Metrics', link: './evnt-eventing-metrics' },
  { text: 'Tutorials', link: './tutorials/evnt-01-prerequisites' },
  { text: 'Create Subscription Subscribing to Multiple Event Types', link: './tutorials/evnt-02-subs-with-multiple-filters' },
  { text: 'Event Name Cleanup in Subscriptions', link: './tutorials/evnt-03-type-cleanup' },
  { text: 'Changing Events Max-In-Flight in Subscriptions', link: './tutorials/evnt-04-change-max-in-flight-in-sub' },
  { text: 'Publish Legacy Events Using Kyma Eventing', link: './tutorials/evnt-05-send-legacy-events' },
  { text: 'Resources', link: './resources/README', collapsed: true, items: [
    { text: 'Subscription CR', link: './resources/evnt-cr-subscription' }
    ] },
  { text: 'Troubleshooting', link: './troubleshooting/README', collapsed: true, items: [
    { text: 'Kyma Eventing - Basic Diagnostics', link: './troubleshooting/evnt-01-eventing-troubleshooting' },
    { text: 'NATS JetStream Backend Troubleshooting', link: './troubleshooting/evnt-02-jetstream-troubleshooting' },
    { text: 'Subscriber Receives Irrelevant Events', link: './troubleshooting/evnt-03-type-collision' },
    { text: 'Eventing Backend Stopped Receiving Events Due To Full Storage', link: './troubleshooting/evnt-04-free-jetstream-storage' },
    { text: 'Published Events Are Pending in the Stream', link: './troubleshooting/evnt-05-fix-pending-messages' }
    ] }
];
