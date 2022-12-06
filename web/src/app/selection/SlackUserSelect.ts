import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: SlackUserSearchOptions) {
    slackUsers(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query ($id: ID!) {
    slackUser(id: $id) {
      id
      name
    }
  }
`

export const SlackChannelSelect = makeQuerySelect('SlackUserSelect', {
  query,
  valueQuery,
})
