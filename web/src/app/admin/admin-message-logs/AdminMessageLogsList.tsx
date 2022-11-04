import React from 'react'
import { DebugMessage } from '../../../schema'
import AdminMessageLogCard from './AdminMessageLogCard'
import { Button, Grid, Typography } from '@mui/material'

interface Props {
  debugMessages: DebugMessage[]
  selectedLog: DebugMessage | null
  onSelect: (debugMessage: DebugMessage) => void
  onLoadMore: () => void
  hasMore: boolean
}

export default function AdminMessageLogsList(props: Props): JSX.Element {
  const { debugMessages, selectedLog, onSelect, hasMore, onLoadMore } = props

  return (
    <Grid
      aria-label='Message Logs List'
      container
      direction='column'
      spacing={2}
    >
      {debugMessages.map((msg) => (
        <Grid key={msg.id} item xs={12}>
          <AdminMessageLogCard
            debugMessage={msg}
            selected={selectedLog?.id === msg.id}
            onSelect={() => onSelect(msg)}
          />
        </Grid>
      ))}

      <Grid
        item
        xs={12}
        sx={{
          display: 'flex',
          justifyContent: 'center',
        }}
      >
        {hasMore ? (
          <Button variant='contained' onClick={onLoadMore}>
            Show more
          </Button>
        ) : (
          <Typography color='textSecondary' variant='body2'>
            Displaying all results.
          </Typography>
        )}
      </Grid>
    </Grid>
  )
}