import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as meetingsApi from '../api/meetings';
import { CreateMeetingInput, VoteInput } from '../api/types';

export function useMeetingsList() {
  return useQuery({
    queryKey: ['meetings'],
    queryFn: meetingsApi.listMeetings,
  });
}

export function useAllMeetings() {
  return useQuery({
    queryKey: ['meetings', 'all'],
    queryFn: meetingsApi.listAllMeetings,
  });
}

export function useMeetingDetail(id: string) {
  return useQuery({
    queryKey: ['meetings', id],
    queryFn: () => meetingsApi.getMeeting(id),
    enabled: !!id,
  });
}

export function useCreateMeeting() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateMeetingInput) => meetingsApi.createMeeting(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['meetings'] });
    },
  });
}

export function useDeleteMeeting() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => meetingsApi.deleteMeeting(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['meetings'] });
    },
  });
}

export function useConfirmMeeting() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, timeSlotId }: { id: string; timeSlotId?: string }) =>
      meetingsApi.confirmMeeting(id, timeSlotId),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['meetings', id] });
      queryClient.invalidateQueries({ queryKey: ['meetings'] });
    },
  });
}

export function useVote() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: VoteInput }) =>
      meetingsApi.vote(id, input),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['meetings', id] });
    },
  });
}

export function useUpdateRSVP() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'accepted' | 'declined' }) =>
      meetingsApi.updateRSVP(id, status),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['meetings', id] });
    },
  });
}

export function useSetTags() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, tags }: { id: string; tags: string[] }) =>
      meetingsApi.setTags(id, tags),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ['meetings', id] });
      queryClient.invalidateQueries({ queryKey: ['meetings'] });
      queryClient.invalidateQueries({ queryKey: ['tags'] });
    },
  });
}

export function useAllTags() {
  return useQuery({
    queryKey: ['tags'],
    queryFn: meetingsApi.getAllTags,
  });
}
